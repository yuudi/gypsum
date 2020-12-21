package gypsum

import (
	"bytes"
	"encoding/gob"
	zero "github.com/wdvxdr1123/ZeroBot"
	"log"
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

func (h *Rule) Register(name string) (err error) {
	tmpl, err := template.New(name).Parse(h.Response)
	if err != nil {
		log.Printf("模板预处理出错：%s", err)
		return
	}
	switch h.MatcherType {
	case FullMatch:
		zero.OnFullMatchGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	case Keyword:
		zero.OnKeywordGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	case Prefix:
		zero.OnPrefixGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	case Suffix:
		zero.OnSuffixGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	case Command:
		zero.OnCommandGroup(h.Patterns, groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	case Regex:
		zero.OnRegex(h.Patterns[0], groupRule(h.GroupID), userRule(h.UserID)).SetPriority(h.Priority).SetBlock(h.Block).Handle(templateHandler(*tmpl))
	default:
		log.Printf("Unknown type %d", h.MatcherType)
	}
	return
}

func templateHandler(tmpl template.Template) zero.Handler {
	return func(_ *zero.Matcher, event zero.Event, _ zero.State) zero.Response {
		buf := &bytes.Buffer{}
		if err := tmpl.Execute(buf, event); err != nil {
			log.Printf("渲染模板出错：%s", err)
			return zero.FinishResponse
		}
		zero.Send(event, buf.String())
		return zero.FinishResponse
	}
}
