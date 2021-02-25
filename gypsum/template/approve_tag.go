package template

import (
	"fmt"

	"github.com/flosch/pongo2"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var Bot *zero.Ctx

type tagApproveNode struct{}

func (node *tagApproveNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	eventCtx, ok := ctx.Public["_event"]
	if !ok {
		return ctx.Error("cannot approve without event", nil)
	}
	event, ok := eventCtx.(*zero.Event)
	if !ok {
		return ctx.Error(fmt.Sprintf("event type error: %#v", eventCtx), nil)
	}
	if event == nil {
		return ctx.Error("event is nil", nil)
	}
	if event.PostType != "request" {
		return ctx.Error(fmt.Sprintf("cannot approve: event is not a request: %#v", event), nil)
	}
	switch event.RequestType {
	case "friend":
		go Bot.SetFriendAddRequest(event.Flag, true, "")
	case "group":
		go Bot.SetGroupAddRequest(event.Flag, event.SubType, true, "")
	}
	return nil
}

func TagApproveParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	return &tagApproveNode{}, nil
}
