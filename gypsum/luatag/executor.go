package luatag

import (
	"bytes"
	"net/http"

	"github.com/cjoudrey/gluahttp"
	"github.com/flosch/pongo2"
	log "github.com/sirupsen/logrus"
	"github.com/yuin/gopher-lua"
	luaJson "layeh.com/gopher-json"
)

type tagLuaNode struct {
	wrapper *pongo2.NodeWrapper
}

func (node tagLuaNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	b := bytes.NewBuffer(make([]byte, 0, 1024)) // 1 KiB
	if err := node.wrapper.Execute(ctx, b); err != nil {
		return err
	}
	s := b.String()

	L := ctx.Public["_lua"].(*lua.LState)
	if L == nil {
		L = lua.NewState()
		// the close function are called by executor caller

		L.PreloadModule("bot", botModLoader)
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
		L.SetGlobal("write", L.NewFunction(Writer(writer, false)))
		L.SetGlobal("write_safe", L.NewFunction(Writer(writer, true)))
		L.SetGlobal("event", luaEvent)
		ctx.Public["_lua"] = L
	}
	if err := L.DoString(s); err != nil {
		log.Errorf("lua execution error: %s", err)
		return nil
	}
	return nil
}

func TagLuaParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	luaNode := &tagLuaNode{}
	wrapper, _, err := doc.WrapUntilTag("endlua")
	if err != nil {
		return nil, err
	}
	luaNode.wrapper = wrapper
	return luaNode, nil
}
