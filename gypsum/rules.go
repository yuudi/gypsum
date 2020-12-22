package gypsum

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	zero "github.com/wdvxdr1123/ZeroBot"
	"log"
	"strconv"
	"strings"
)

type RuleType int

const (
	FullMatch RuleType = iota
	Keyword
	Prefix
	Suffix
	Command
	Regex
)

type Rule struct {
	GroupID     int64    `json:"group_id"`
	UserID      int64    `json:"user_id"`
	MatcherType RuleType `json:"matcher_type"`
	Patterns    []string `json:"patterns"`
	Response    string   `json:"response"`
	Priority    int      `json:"priority"`
	Block       bool     `json:"block"`
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
	h := &Rule{}
	buffer := bytes.Buffer{}
	buffer.Write(b)
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(h)
	return h, err
}

func groupRule(groupId int64) zero.Rule {
	if groupId == 0 {
		return func(_ *zero.Event, _ zero.State) bool {
			return true
		}
	}
	return func(event *zero.Event, _ zero.State) bool {
		return event.GroupID == groupId
	}
}
func userRule(userId int64) zero.Rule {
	if userId == 0 {
		return func(_ *zero.Event, _ zero.State) bool {
			return true
		}
	}
	return func(event *zero.Event, _ zero.State) bool {
		return event.UserID == userId
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
		zeroMatcher[id] = zero.OnFullMatchGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	case Keyword:
		zeroMatcher[id] = zero.OnKeywordGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	case Prefix:
		zeroMatcher[id] = zero.OnPrefixGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	case Suffix:
		zeroMatcher[id] = zero.OnSuffixGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	case Command:
		zeroMatcher[id] = zero.OnCommandGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
	case Regex:
		zeroMatcher[id] = zero.OnRegex(h.Patterns[0], groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateRuleHandler(*tmpl))
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

func getRules(c *gin.Context) {
	c.JSON(200, rules)
}

func getRuleById(c *gin.Context) {
	ruleIdStr := c.Param("rid")
	ruleId, err := strconv.ParseUint(ruleIdStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such rule",
		})
	} else {
		r, ok := rules[ruleId]
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
	ruleIdStr := c.Param("rid")
	ruleId, err := strconv.ParseUint(ruleIdStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such rule",
		})
	} else {
		_, ok := rules[ruleId]
		if ok {
			delete(rules, ruleId)
			if err := db.Delete(append([]byte("gypsum-rules-"), ToBytes(ruleId)...), nil); err != nil {
				c.JSON(500, gin.H{
					"code":    3001,
					"message": fmt.Sprintf("Server got itself into trouble: %s", err),
				})
				return
			}
			zeroMatcher[ruleId].Delete()
			c.JSON(200, gin.H{
				"code":    0,
				"message": "deleted",
			})
		} else {
			c.JSON(404, gin.H{
				"code":    1000,
				"message": "no such rule",
			})
		}
	}
}

func modifyRule(c *gin.Context) {
	ruleIdStr := c.Param("rid")
	ruleId, err := strconv.ParseUint(ruleIdStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such rule",
		})
		return
	}
	_, ok := rules[ruleId]
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
	matcher := zeroMatcher[ruleId]
	if err := rule.Register(ruleId); err != nil {
		c.JSON(400, gin.H{
			"code":    2001,
			"message": fmt.Sprintf("rule error: %s", err),
		})
		return
	}
	matcher.Delete()
	if err := db.Put(append([]byte("gypsum-rules-"), ToBytes(ruleId)...), v, nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3002,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	rules[ruleId] = rule
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}
