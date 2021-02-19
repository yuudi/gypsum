package gypsum

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	zeroMessage "github.com/wdvxdr1123/ZeroBot/message"
	lua "github.com/yuin/gopher-lua"
)

type testCase struct {
	Event       gjson.Result
	DebugType   string
	MatcherType RuleType
	Pattern     string
	Response    string
}

type responseReceiver struct {
	contents strings.Builder
}

func (r *responseReceiver) ReceiveSend(_ zero.Event, msg interface{}) int64 {
	r.contents.WriteString(msg.(string))
	return 0
}
func (r *responseReceiver) ReceiveLogger(i ...interface{}) {
	r.contents.WriteString(fmt.Sprint(i...))
}

func (r *responseReceiver) String() string {
	return r.contents.String()
}

func (t *testCase) TestMessage() (string, bool, error) {
	var zeroRule zero.Rule
	switch t.MatcherType {
	case FullMatch:
		zeroRule = zero.FullMatchRule(t.Pattern)
	case Keyword:
		zeroRule = zero.KeywordRule(t.Pattern)
	case Prefix:
		zeroRule = zero.PrefixRule(t.Pattern)
	case Suffix:
		zeroRule = zero.SuffixRule(t.Pattern)
	case Command:
		zeroRule = zero.CommandRule(t.Pattern)
	case Regex:
		if err := checkRegex(t.Pattern); err != nil {
			return "", false, errors.New("正则语法错误：" + err.Error())
		}
		zeroRule = zero.RegexRule(t.Pattern)
	default:
		return "", false, errors.New(fmt.Sprintf("Unknown type %#v", t.MatcherType))
	}
	var event zero.Event
	var state zero.State = make(map[string]interface{})
	err := jsoniter.UnmarshalFromString(t.Event.String(), &event)
	if err != nil {
		return "", false, errors.New("json解析出错：" + err.Error())
	}
	event.Message = zeroMessage.ParseMessageFromString(t.Event.Get("message").String())
	event.RawEvent = t.Event
	log.Debug(event.Message.CQString())
	matched := zeroRule(&event, state)
	if !matched {
		return "", false, nil
	}
	tmpl, err := pongo2.FromString(t.Response)
	if err != nil {
		return "", true, errors.New("模板预处理出错：" + err.Error())
	}
	var receiver responseReceiver
	handler := templateRuleHandler(*tmpl, receiver.ReceiveSend, receiver.ReceiveLogger)
	handler(nil, event, state)
	return receiver.String(), true, nil
}

func (t *testCase) TestNotice() (string, error) {
	tmpl, err := pongo2.FromString(t.Response)
	if err != nil {
		return "", errors.New("模板预处理出错：" + err.Error())
	}
	var event zero.Event
	var state zero.State
	event.RawEvent = t.Event
	var receiver responseReceiver
	handler := templateTriggerHandler(*tmpl, receiver.ReceiveSend, receiver.ReceiveLogger)
	handler(nil, event, state)
	return receiver.String(), nil
}

func (t *testCase) TestTemplate() (string, error) {
	tmpl, err := pongo2.FromString(t.Response)
	if err != nil {
		return "", errors.New("模板预处理出错：" + err.Error())
	}
	var luaState *lua.LState
	defer func() {
		if luaState != nil {
			luaState.Close()
		}
	}()
	msg, err := tmpl.Execute(pongo2.Context{
		"_lua": luaState,
	})
	if err != nil {
		return "", errors.New("渲染模板出错：" + err.Error())
	}
	msg = strings.TrimSpace(msg)
	return msg, nil
}

func (t *testCase) RunTest() (string, bool, error) {
	switch t.DebugType {
	case "message":
		return t.TestMessage()
	case "notice":
		r, e := t.TestNotice()
		return r, true, e
	case "schedule":
		r, e := t.TestTemplate()
		return r, true, e
	default:
		err := errors.New("unknown debug_type: " + t.DebugType)
		return "", false, err
	}
}

func userTest(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("read error: %s", err),
		})
		return
	}
	req := gjson.ParseBytes(body).Map()
	var t = testCase{
		Event:       req["event"],
		DebugType:   req["debug_type"].String(),
		MatcherType: RuleType(req["matcher_type"].Int()),
		Pattern:     req["pattern"].String(),
		Response:    req["response"].String(),
	}
	reply, matched, err := t.RunTest()
	if err != nil {
		c.JSON(200, gin.H{
			"code":    2800,
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code":    0,
		"matched": matched,
		"reply":   reply,
	})
}
