package gypsum

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb/util"
	zero "github.com/wdvxdr1123/ZeroBot"
	lua "github.com/yuin/gopher-lua"
)

type RuleType int
type MessageType uint32

const (
	FullMatch RuleType = iota
	Keyword
	Prefix
	Suffix
	Command
	Regex
)

const (
	FriendMessage MessageType = 1 << iota
	GroupTmpMessage
	OtherTmpMessage
	OfficialMessage
	GroupNormalMessage
	GroupAnonymousMessage
	GroupNoticeMessage
	DiscussMessage

	NoMessage      MessageType = 0
	AllMessage     MessageType = 0xffffffff
	PrivateMessage             = FriendMessage | GroupTmpMessage | OtherTmpMessage
	GroupMessage               = GroupNormalMessage | GroupAnonymousMessage | GroupNoticeMessage
)

var messageTypeTable = map[string]MessageType{
	"group":    GroupMessage,
	"private":  PrivateMessage,
	"discuss":  DiscussMessage,
	"official": OfficialMessage,
}

type Rule struct {
	DisplayName string      `json:"display_name"`
	Active      bool        `json:"active"`
	MessageType MessageType `json:"message_type"`
	GroupID     int64       `json:"group_id"`
	UserID      int64       `json:"user_id"`
	MatcherType RuleType    `json:"matcher_type"`
	Patterns    []string    `json:"patterns"`
	OnlyAtMe    bool        `json:"only_at_me"`
	Response    string      `json:"response"`
	Priority    int         `json:"priority"`
	Block       bool        `json:"block"`
}

var (
	rules       map[uint64]Rule
	zeroMatcher map[uint64]*zero.Matcher
)

func (h *Rule) ToBytes() ([]byte, error) {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(h); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func RuleFromBytes(b []byte) (*Rule, error) {
	r := &Rule{
		Active:      true,
		MessageType: AllMessage,
		GroupID:     0,
		UserID:      0,
		MatcherType: FullMatch,
		Patterns:    []string{},
		OnlyAtMe:    false,
		Response:    "",
		Priority:    50,
		Block:       true,
	}
	buffer := bytes.Buffer{}
	buffer.Write(b)
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(r)
	return r, err
}

func typeRule(acceptType MessageType) zero.Rule {
	return func(event *zero.Event, _ zero.State) bool {
		msgType, ok := messageTypeTable[event.MessageType]
		if !ok {
			log.Warnf("未知的消息类型：%s", event.MessageType)
			return false
		}
		return (msgType & acceptType) != NoMessage
	}
}

func groupRule(groupID int64) zero.Rule {
	if groupID == 0 {
		return func(_ *zero.Event, _ zero.State) bool {
			return true
		}
	}
	return func(event *zero.Event, _ zero.State) bool {
		return event.GroupID == groupID
	}
}
func userRule(userID int64) zero.Rule {
	if userID == 0 {
		return func(_ *zero.Event, _ zero.State) bool {
			return true
		}
	}
	return func(event *zero.Event, _ zero.State) bool {
		return event.UserID == userID
	}
}

func (h *Rule) Register(id uint64) error {
	if !h.Active {
		return nil
	}
	tmpl, err := pongo2.FromString(h.Response)
	if err != nil {
		log.Errorf("模板预处理出错：%s", err)
		return err
	}
	rules := []zero.Rule{typeRule(h.MessageType)}
	if h.GroupID != 0 {
		rules = append(rules, groupRule(h.GroupID))
	}
	if h.UserID != 0 {
		rules = append(rules, userRule(h.UserID))
	}
	if h.OnlyAtMe {
		rules = append(rules, zero.OnlyToMe)
	}
	var msgRule zero.Rule
	switch h.MatcherType {
	case FullMatch:
		msgRule = zero.FullMatchRule(h.Patterns...)
	case Keyword:
		msgRule = zero.KeywordRule(h.Patterns...)
	case Prefix:
		msgRule = zero.PrefixRule(h.Patterns...)
	case Suffix:
		msgRule = zero.SuffixRule(h.Patterns...)
	case Command:
		msgRule = zero.CommandRule(h.Patterns...)
	case Regex:
		msgRule = zero.RegexRule(h.Patterns[0])
	default:
		log.Errorf("Unknown type %#v", h.MatcherType)
		return errors.New(fmt.Sprintf("Unknown type %#v", h.MatcherType))
	}
	zeroMatcher[id] = zero.OnMessage(append(rules, msgRule)...).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	return nil
}

func templateRuleHandler(tmpl pongo2.Template) zero.Handler {
	return func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
		var luaState *lua.LState
		defer func() {
			if luaState != nil {
				luaState.Close()
			}
		}()
		reply, err := tmpl.Execute(pongo2.Context{
			"matcher": matcher,
			"state":   state,
			"event": func() interface{} {
				e := make(map[string]interface{})
				if err := jsoniter.UnmarshalFromString(event.RawEvent.Raw, &e); err != nil {
					log.Errorf("error when decode event json: %s", err)
				}
				return e
			},
			"json_event": &event.RawEvent.Raw,
			"at_sender": func() string {
				if event.GroupID == 0 {
					log.Warnf("cannot at sender in event %s/%s", event.PostType, event.SubType)
					return ""
				}
				return fmt.Sprintf("[CQ:at,qq=%d]", event.UserID)
			},
			"_lua": luaState,
		})
		if err != nil {
			log.Errorf("error when rendering template：%s", err)
			return zero.FinishResponse
		}
		reply = strings.TrimSpace(reply)
		if reply != "" {
			zero.Send(event, reply)
		}
		return zero.FinishResponse
	}
}

func loadRules() {
	rules = make(map[uint64]Rule)
	zeroMatcher = make(map[uint64]*zero.Matcher)
	iter := db.NewIterator(util.BytesPrefix([]byte("gypsum-rules-")), nil)
	defer func() {
		iter.Release()
		if err := iter.Error(); err != nil {
			log.Errorf("载入数据错误：%s", err)
		}
	}()
	for iter.Next() {
		key := ToUint(iter.Key()[13:])
		value := iter.Value()
		r, e := RuleFromBytes(value)
		if e != nil {
			log.Errorf("无法加载规则%d：%s", key, e)
			continue
		}
		rules[key] = *r
		if e := r.Register(key); e != nil {
			log.Errorf("无法注册规则%d：%s", key, e)
			continue
		}
	}
}

func checkRegex(pattern string) error {
	_, err := regexp.Compile(pattern)
	return err
}

func checkTemplate(template string) error {
	_, err := pongo2.FromString(template)
	return err
}

func getRules(c *gin.Context) {
	c.JSON(200, rules)
}

func getRuleByID(c *gin.Context) {
	ruleIDStr := c.Param("rid")
	ruleID, err := strconv.ParseUint(ruleIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such rule",
		})
	} else {
		r, ok := rules[ruleID]
		if ok {
			c.JSON(200, r)
		} else {
			c.JSON(404, gin.H{
				"code":    1000,
				"message": "no such rule",
			})
		}
	}
}

func createRule(c *gin.Context) {
	var rule Rule
	if err := c.BindJSON(&rule); err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	if rule.MatcherType == Regex {
		if len(rule.Patterns) != 1 {
			c.JSON(422, gin.H{
				"code":    2001,
				"message": fmt.Sprintf("regex mather can only accept one pattern"),
			})
			return
		}
		if err := checkRegex(rule.Patterns[1]); err != nil {
			c.JSON(422, gin.H{
				"code":    2002,
				"message": fmt.Sprintf("cannot compile regex pattern: %s", err),
			})
			return
		}
	}
	if err := checkTemplate(rule.Response); err != nil {
		c.JSON(422, gin.H{
			"code":    2041,
			"message": fmt.Sprintf("template error: %s", err),
		})
		return
	}
	cursor++
	if err := db.Put([]byte("gypsum-$meta-cursor"), U64ToBytes(cursor), nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	v, err := rule.ToBytes()
	if err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	if err := rule.Register(cursor); err != nil {
		c.JSON(400, gin.H{
			"code":    2001,
			"message": fmt.Sprintf("rule error: %s", err),
		})
		return
	}
	if err := db.Put(append([]byte("gypsum-rules-"), U64ToBytes(cursor)...), v, nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	rules[cursor] = rule
	c.JSON(201, gin.H{
		"code":    0,
		"message": "ok",
		"rule_id": cursor,
	})
	return
}

func deleteRule(c *gin.Context) {
	ruleIDStr := c.Param("rid")
	ruleID, err := strconv.ParseUint(ruleIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such rule",
		})
		return
	}
	oldRule, ok := rules[ruleID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such rule",
		})
		return
	}
	delete(rules, ruleID)
	if err := db.Delete(append([]byte("gypsum-rules-"), U64ToBytes(ruleID)...), nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3001,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	if oldRule.Active {
		zeroMatcher[ruleID].Delete()
	}
	c.JSON(200, gin.H{
		"code":    0,
		"message": "deleted",
	})
	return
}

func modifyRule(c *gin.Context) {
	ruleIDStr := c.Param("rid")
	ruleID, err := strconv.ParseUint(ruleIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such rule",
		})
		return
	}
	oldRule, ok := rules[ruleID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    100,
			"message": "no such rule",
		})
		return
	}
	var newRule Rule
	if err := c.BindJSON(&newRule); err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	if newRule.MatcherType == Regex {
		if len(newRule.Patterns) != 1 {
			c.JSON(422, gin.H{
				"code":    2001,
				"message": fmt.Sprintf("regex mather can only accept one pattern"),
			})
			return
		}
		if err := checkRegex(newRule.Patterns[1]); err != nil {
			c.JSON(422, gin.H{
				"code":    2002,
				"message": fmt.Sprintf("cannot compile regex pattern: %s", err),
			})
			return
		}
	}
	if err := checkTemplate(newRule.Response); err != nil {
		c.JSON(422, gin.H{
			"code":    2041,
			"message": fmt.Sprintf("template error: %s", err),
		})
		return
	}
	v, err := newRule.ToBytes()
	if err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	oldMatcher, ok := zeroMatcher[ruleID]
	if err := newRule.Register(ruleID); err != nil {
		c.JSON(400, gin.H{
			"code":    2001,
			"message": fmt.Sprintf("rule error: %s", err),
		})
		return
	}
	if oldRule.Active {
		if !ok {
			c.JSON(500, gin.H{
				"code":    7012,
				"message": "error when delete old rule: matcher not found",
			})
			return
		}
		oldMatcher.Delete()
	}
	if err := db.Put(append([]byte("gypsum-rules-"), U64ToBytes(ruleID)...), v, nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3002,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	rules[ruleID] = newRule
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}
