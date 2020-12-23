package gypsum

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	"github.com/syndtr/goleveldb/leveldb/util"
	zero "github.com/wdvxdr1123/ZeroBot"
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
	Active      bool        `json:"active"`
	MessageType MessageType `json:"message_type"`
	GroupID     int64       `json:"group_id"`
	UserID      int64       `json:"user_id"`
	MatcherType RuleType    `json:"matcher_type"`
	Patterns    []string    `json:"patterns"`
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
			log.Printf("未知的消息类型：%s", event.MessageType)
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
	tmpl, err := pongo2.FromString(h.Response)
	if err != nil {
		log.Printf("模板预处理出错：%s", err)
		return err
	}
	switch h.MatcherType {
	case FullMatch:
		zeroMatcher[id] = zero.OnFullMatchGroup(h.Patterns, typeRule(h.MessageType), groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	case Keyword:
		zeroMatcher[id] = zero.OnKeywordGroup(h.Patterns, typeRule(h.MessageType), groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	case Prefix:
		zeroMatcher[id] = zero.OnPrefixGroup(h.Patterns, typeRule(h.MessageType), groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	case Suffix:
		zeroMatcher[id] = zero.OnSuffixGroup(h.Patterns, typeRule(h.MessageType), groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	case Command:
		zeroMatcher[id] = zero.OnCommandGroup(h.Patterns, typeRule(h.MessageType), groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	case Regex:
		zeroMatcher[id] = zero.OnRegex(h.Patterns[0], typeRule(h.MessageType), groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	default:
		log.Printf("Unknown type %d", h.MatcherType)
	}
	return nil
}

func templateRuleHandler(tmpl pongo2.Template) zero.Handler {
	return func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
		reply, err := tmpl.Execute(pongo2.Context{
			"matcher": matcher,
			"event":   event,
			"state":   state,
		})
		if err != nil {
			log.Printf("渲染模板出错：%s", err)
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
			log.Printf("载入数据错误：%s", err)
		}
	}()
	for iter.Next() {
		key := ToUint(iter.Key()[13:])
		value := iter.Value()
		r, e := RuleFromBytes(value)
		if e != nil {
			log.Printf("无法加载规则%d：%s", key, e)
			continue
		}
		rules[key] = *r
		if e := r.Register(key); e != nil {
			log.Printf("无法注册规则%d：%s", key, e)
			continue
		}
	}
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
	cursor++
	if err := db.Put([]byte("gypsum-$meta-cursor"), ToBytes(cursor), nil); err != nil {
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
	if err := db.Put(append([]byte("gypsum-rules-"), ToBytes(cursor)...), v, nil); err != nil {
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
	_, ok := rules[ruleID]
	if ok {
		delete(rules, ruleID)
		if err := db.Delete(append([]byte("gypsum-rules-"), ToBytes(ruleID)...), nil); err != nil {
			c.JSON(500, gin.H{
				"code":    3001,
				"message": fmt.Sprintf("Server got itself into trouble: %s", err),
			})
			return
		}
		zeroMatcher[ruleID].Delete()
		c.JSON(200, gin.H{
			"code":    0,
			"message": "deleted",
		})
		return
	}
	c.JSON(404, gin.H{
		"code":    1000,
		"message": "no such rule",
	})

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
	_, ok := rules[ruleID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    100,
			"message": "no such rule",
		})
		return
	}
	var rule Rule
	if err := c.BindJSON(&rule); err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
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
	matcher := zeroMatcher[ruleID]
	if err := rule.Register(ruleID); err != nil {
		c.JSON(400, gin.H{
			"code":    2001,
			"message": fmt.Sprintf("rule error: %s", err),
		})
		return
	}
	matcher.Delete()
	if err := db.Put(append([]byte("gypsum-rules-"), ToBytes(ruleID)...), v, nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3002,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	rules[ruleID] = rule
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}
