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
	GroupsID    []int64     `json:"groups_id"`
	UsersID     []int64     `json:"users_id"`
	MatcherType RuleType    `json:"matcher_type"`
	Patterns    []string    `json:"patterns"`
	OnlyAtMe    bool        `json:"only_at_me"`
	Response    string      `json:"response"`
	Priority    int         `json:"priority"`
	Block       bool        `json:"block"`
	ParentGroup uint64      `json:"-"`
}

var (
	rules       map[uint64]*Rule
	zeroMatcher map[uint64]*zero.Matcher
)

func (r *Rule) ToBytes() ([]byte, error) {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(r); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func RuleFromBytes(b []byte) (*Rule, error) {
	r := &Rule{
		DisplayName: "",
		Active:      true,
		MessageType: AllMessage,
		GroupsID:    []int64{},
		UsersID:     []int64{},
		MatcherType: FullMatch,
		Patterns:    []string{},
		OnlyAtMe:    false,
		Response:    "",
		Priority:    50,
		Block:       true,
		ParentGroup: 0,
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

func groupsRule(groupsID []int64) zero.Rule {
	if len(groupsID) == 0 {
		return func(_ *zero.Event, _ zero.State) bool {
			return true
		}
	}
	return func(event *zero.Event, _ zero.State) bool {
		for _, i := range groupsID {
			if i == event.GroupID {
				return true
			}
		}
		return false
	}
}

func usersRule(usersID []int64) zero.Rule {
	if len(usersID) == 0 {
		return func(_ *zero.Event, _ zero.State) bool {
			return true
		}
	}
	return func(event *zero.Event, _ zero.State) bool {
		for _, i := range usersID {
			if i == event.UserID {
				return true
			}
		}
		return false
	}
}

func (r *Rule) Register(id uint64) error {
	if !r.Active {
		return nil
	}
	tmpl, err := pongo2.FromString(r.Response)
	if err != nil {
		log.Errorf("模板预处理出错：%s", err)
		return err
	}
	rules := []zero.Rule{typeRule(r.MessageType)}
	if len(r.GroupsID) != 0 {
		rules = append(rules, groupsRule(r.GroupsID))
	}
	if len(r.UsersID) != 0 {
		rules = append(rules, usersRule(r.UsersID))
	}
	if r.OnlyAtMe {
		rules = append(rules, zero.OnlyToMe)
	}
	var msgRule zero.Rule
	switch r.MatcherType {
	case FullMatch:
		msgRule = zero.FullMatchRule(r.Patterns...)
	case Keyword:
		msgRule = zero.KeywordRule(r.Patterns...)
	case Prefix:
		msgRule = zero.PrefixRule(r.Patterns...)
	case Suffix:
		msgRule = zero.SuffixRule(r.Patterns...)
	case Command:
		msgRule = zero.CommandRule(r.Patterns...)
	case Regex:
		if len(r.Patterns) == 0 {
			msgRule = func(_ *zero.Event, _ zero.State) bool {
				return false
			}
		} else {
			msgRule = zero.RegexRule(r.Patterns[0])
		}
	default:
		log.Errorf("Unknown type %#v", r.MatcherType)
		return errors.New(fmt.Sprintf("Unknown type %#v", r.MatcherType))
	}
	zeroMatcher[id] = zero.OnMessage(append(rules, msgRule)...).SetPriority(r.Priority).SetBlock(r.Block).Handle(templateRuleHandler(*tmpl, zero.Send, log.Error))
	return nil
}

func templateRuleHandler(tmpl pongo2.Template, send func(event zero.Event, msg interface{}) int64, errLogger func(...interface{})) zero.Handler {
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
			"group_ban": func(duration interface{}) {
				if event.GroupID == 0 {
					log.Warnf("cannot ban sender in event %s/%s", event.PostType, event.SubType)
					return
				}
				d, err := AnyToInt64(duration)
				if err != nil {
					log.Warnf("cannot convert %#v to int64", duration)
					return
				}
				zero.SetGroupBan(event.GroupID, event.UserID, d)
			},
			"_event": &event,
			"_lua":   luaState,
		})
		if err != nil {
			errLogger("渲染模板出错：" + err.Error())
			return zero.FinishResponse
		}
		reply = strings.TrimSpace(reply)
		if reply != "" {
			send(event, reply)
		}
		return zero.FinishResponse
	}
}

func loadRules() {
	rules = make(map[uint64]*Rule)
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
		rules[key] = r
		if e := r.Register(key); e != nil {
			log.Errorf("无法注册规则%d：%s", key, e)
			continue
		}
	}
}

func (r *Rule) SaveToDB(idx uint64) error {
	v, err := r.ToBytes()
	if err != nil {
		return err
	}
	return db.Put(append([]byte("gypsum-rules-"), U64ToBytes(idx)...), v, nil)
}

func checkRegex(pattern string) error {
	_, err := regexp.Compile(pattern)
	return err
}

func checkTemplate(template string) error {
	_, err := pongo2.FromString(template)
	return err
}

func (r *Rule) GetParentID() uint64 {
	return r.ParentGroup
}

func (r *Rule) GetDisplayName() string {
	return r.DisplayName
}

func (r *Rule) NewParent(selfID, parentID uint64) error {
	v, err := r.ToBytes()
	if err != nil {
		return err
	}
	r.ParentGroup = parentID
	err = db.Put(append([]byte("gypsum-rules-"), U64ToBytes(selfID)...), v, nil)
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
	parentStr := c.Param("gid")
	var parentID uint64
	if len(parentStr) == 0 {
		parentID = 0
	} else {
		var err error
		parentID, err = strconv.ParseUint(parentStr, 10, 64)
		if err != nil {
			c.JSON(404, gin.H{
				"code":    1000,
				"message": "no such group",
			})
			return
		}
	}
	parentGroup, ok := groups[parentID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "group not found",
		})
		return
	}
	rule.ParentGroup = parentID
	// syntax check
	if rule.MatcherType == Regex {
		if len(rule.Patterns) != 1 {
			c.JSON(422, gin.H{
				"code":    2001,
				"message": fmt.Sprintf("regex mather can only accept one pattern"),
			})
			return
		}
		if err := checkRegex(rule.Patterns[0]); err != nil {
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
	// save
	itemCursor++
	cursor := itemCursor
	parentGroup.Items = append(parentGroup.Items, Item{
		ItemType:    RuleItem,
		DisplayName: rule.DisplayName,
		ItemID:      cursor,
	})
	if err := parentGroup.SaveToDB(parentID); err != nil {
		log.Error(err)
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
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
	rules[cursor] = &rule
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
	// remove self from parent
	if err := DeleteFromParent(oldRule.ParentGroup, ruleID); err != nil {
		log.Errorf("error when delete group %d from parent group %d: %s", ruleID, oldRule.ParentGroup, err)
	}
	// remove self from database
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
	// check new rule syntax
	if newRule.MatcherType == Regex {
		if len(newRule.Patterns) != 1 {
			c.JSON(422, gin.H{
				"code":    2001,
				"message": fmt.Sprintf("regex mather can only accept one pattern"),
			})
			return
		}
		if err := checkRegex(newRule.Patterns[0]); err != nil {
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
	newRule.ParentGroup = oldRule.ParentGroup
	if oldRule.Active {
		oldMatcher, ok := zeroMatcher[ruleID]
		if !ok {
			c.JSON(500, gin.H{
				"code":    7012,
				"message": "error when delete old rule: matcher not found",
			})
			return
		}
		oldMatcher.Delete()
	}
	if err := newRule.Register(ruleID); err != nil {
		c.JSON(400, gin.H{
			"code":    2001,
			"message": fmt.Sprintf("rule error: %s", err),
		})
		return
	}
	if err := newRule.SaveToDB(ruleID); err != nil {
		c.JSON(500, gin.H{
			"code":    3002,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	rules[ruleID] = &newRule
	if newRule.DisplayName != oldRule.DisplayName {
		if err = ChangeNameForParent(newRule.ParentGroup, ruleID, newRule.DisplayName); err != nil {
			log.Errorf("error when change rule %d from parent group %d: %s", ruleID, newRule.ParentGroup, err)
		}
	}
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}
