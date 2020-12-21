package gypsum

import (
	"bytes"
	"encoding/gob"
	"github.com/Masterminds/sprig"
	zero "github.com/wdvxdr1123/ZeroBot"
	"log"
	"strconv"
	"text/template"
)

type MatcherType int

const (
	FullMatch MatcherType = iota
	Keyword
	Prefix
	Suffix
	Command
	Regex
)

type Rule struct {
	GroupID     int64       `json:"group_id"`
	UserID      int64       `json:"user_id"`
	MatcherType MatcherType `json:"matcher_type"`
	Patterns    []string    `json:"patterns"`
	Response    string      `json:"response"`
	Priority    int         `json:"priority"`
	Block       bool        `json:"block"`
}

var rules map[uint64]Rule
var zeroMatcher map[uint64]*zero.Matcher

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
	return func(event *zero.Event, _ zero.State) bool {
		if groupId == 0 {
			return true
		}
		return event.GroupID == groupId
	}
}
func userRule(userId int64) zero.Rule {
	return func(event *zero.Event, _ zero.State) bool {
		if userId == 0 {
			return true
		}
		return event.UserID == userId
	}
}

func (h *Rule) Register(id uint64) (err error) {
	name:="rule" + strconv.FormatUint(cursor, 10)
	tmpl, err := template.New(name).Funcs(sprig.TxtFuncMap()).Parse(h.Response)
	if err != nil {
		log.Printf("模板预处理出错：%s", err)
		return
	}
	switch h.MatcherType {
	case FullMatch:
		zeroMatcher[id] = zero.OnFullMatchGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	case Keyword:
		zeroMatcher[id] = zero.OnKeywordGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	case Prefix:
		zeroMatcher[id] = zero.OnPrefixGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	case Suffix:
		zeroMatcher[id] = zero.OnSuffixGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	case Command:
		zeroMatcher[id] = zero.OnCommandGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	case Regex:
		zeroMatcher[id] = zero.OnRegex(h.Patterns[0], groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	default:
		log.Printf("Unknown type %d", h.MatcherType)
	}
	return
}

func templateHandler(tmpl template.Template) zero.Handler {
	return func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
		buf := &bytes.Buffer{}
		if err := tmpl.Execute(buf, executor{*matcher, event, state}); err != nil {
			log.Printf("渲染模板出错：%s", err)
			return zero.FinishResponse
		}
		zero.Send(event, buf.String())
		return zero.FinishResponse
	}
}
