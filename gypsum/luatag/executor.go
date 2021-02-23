package luatag

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cjoudrey/gluahttp"
	"github.com/flosch/pongo2"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/yuin/gopher-lua"
	luaJson "layeh.com/gopher-json"
)

type tagLuaNode struct {
	Timeout pongo2.IEvaluator
	wrapper *pongo2.NodeWrapper
}

func (node tagLuaNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	var timeoutDuration time.Duration
	if node.Timeout == nil {
		timeoutDuration = 300 * time.Second
	} else {
		// parse time out argument
		v, err := node.Timeout.Evaluate(ctx)
		if err != nil {
			return err
		}
		valueString := v.String()
		var e error
		seconds, e := strconv.ParseInt(valueString, 10, 32)
		if e == nil {
			// all digits
			timeoutDuration = time.Duration(seconds) * time.Second
		} else {
			timeoutDuration, e = time.ParseDuration(valueString)
			if e != nil {
				return ctx.Error("unknown time duration: "+e.Error(), nil)
			}
		}
	}
	b := bytes.NewBuffer(make([]byte, 0, 1024)) // 1 KiB
	if err := node.wrapper.Execute(ctx, b); err != nil {
		return err
	}
	s := b.String()

	L := ctx.Public["_lua"].(*lua.LState)
	if L == nil {
		L = lua.NewState()
		var metaEvent *zero.Event
		metaEventInterface, ok := ctx.Public["_event"]
		if ok {
			metaEvent = metaEventInterface.(*zero.Event)
		} else {
			metaEvent = nil
		}

		L.PreloadModule("bot", botModLoaderFunc(metaEvent))
		L.PreloadModule("database", dbLoader)
		L.PreloadModule("json", luaJson.Loader)
		L.PreloadModule("http", gluahttp.NewHttpModule(&http.Client{}).Loader)
		var luaEvent lua.LValue
		event, ok := ctx.Public["json_event"]
		if !ok {
			luaEvent = lua.LNil
		} else {
			var err error
			luaEvent, err = luaJson.Decode(L, []byte(*event.(*string)))
			if err != nil {
				log.Errorf("lua execution error: cannot resume lua event from pongo2 context")
				return nil
			}
		}
		var luaState lua.LValue
		state, ok := ctx.Public["state"]
		if !ok {
			luaState = lua.LNil
		} else {
			luaState = L.NewTable()
			for k, i := range state.(zero.State) {
				switch v := i.(type) {
				case string:
					L.SetField(luaState, k, lua.LString(v))
				case []string:
					list := L.NewTable()
					for _, s := range v {
						list.Append(lua.LString(s))
					}
					L.SetField(luaState, k, list)
				default:
					log.Warnf("unknown type in state: %#v", v)
				}
			}
		}
		L.SetGlobal("write", L.NewFunction(Writer(writer, false)))
		L.SetGlobal("write_safe", L.NewFunction(Writer(writer, true)))
		L.SetGlobal("sleep", L.NewFunction(luaSleep))
		L.SetGlobal("res", L.NewFunction(resFunc))
		L.SetGlobal("event", luaEvent)
		L.SetGlobal("state", luaState)
		ctx.Public["_lua"] = L
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()
	L.SetContext(timeoutContext)
	if err := L.DoString(s); err != nil {
		return ctx.Error(fmt.Sprintf("lua execution error: %s", err), nil)
	}
	return nil
}

func TagLuaParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	luaNode := &tagLuaNode{}
	keyToken := arguments.MatchType(pongo2.TokenIdentifier)
	if keyToken == nil {
		luaNode.Timeout = nil
	} else if keyToken.Val != "timeout" {
		return nil, arguments.Error("unknown keyword", keyToken)
	} else {
		if arguments.Match(pongo2.TokenSymbol, "=") == nil {
			return nil, arguments.Error("Expected '='.", nil)
		}
		valueExpr, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}
		luaNode.Timeout = valueExpr
	}
	wrapper, _, err := doc.WrapUntilTag("endlua", "end_lua")
	if err != nil {
		return nil, err
	}
	luaNode.wrapper = wrapper
	return luaNode, nil
}

var resFunc lua.LGFunction

func SetResFunc(fn func(string) string) {
	resFunc = func(L *lua.LState) int {
		res := L.ToString(1)
		uri := fn(res)
		L.Push(lua.LString(uri))
		return 1
	}
}
