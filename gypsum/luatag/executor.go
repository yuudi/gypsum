package luatag

import (
	"bytes"
	"log"

	"github.com/flosch/pongo2"
	"github.com/yuin/gopher-lua"
)

type tagLuaNode struct {
	wrapper *pongo2.NodeWrapper
}

func (node tagLuaNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	b := bytes.NewBuffer(make([]byte, 0, 1024)) // 1 KiB
	err := node.wrapper.Execute(ctx, b)
	if err != nil {
		return err
	}
	s := b.String()

	L := lua.NewState()
	var luaVar *lua.LTable
	c, ok := ctx.Shared["luavar"]
	if !ok {
		luaVar = L.NewTable()
		if ctx.Shared == nil {
			ctx.Shared = make(pongo2.Context)
		}
		ctx.Shared["luavar"] = luaVar
	} else {
		luaVar, ok = c.(*lua.LTable)
		if !ok {
			log.Printf("lua execution error: cannot resume lua context from pongo2 context")
			return nil
		}
	}
	L.SetGlobal("write", L.NewFunction(Writer(writer)))
	L.SetGlobal("var", luaVar)
	defer L.Close()
	if err := L.DoString(s); err != nil {
		log.Printf("lua execution error: %s", err)
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
