package template

import (
	"fmt"

	"github.com/flosch/pongo2"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type tagWithdrawNode struct{}

func (node *tagWithdrawNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	eventCtx, ok := ctx.Public["_event"]
	if !ok {
		return ctx.Error("cannot withdraw without event", nil)
	}
	event, ok := eventCtx.(*zero.Event)
	if !ok {
		return ctx.Error(fmt.Sprintf("event type error: %#v", eventCtx), nil)
	}
	if event == nil {
		return ctx.Error("event is nil", nil)
	}
	if event.MessageType != "group" {
		return ctx.Error(fmt.Sprintf("cannot withdraw in event %s/%s", event.PostType, event.DetailType), nil)
	}
	go zero.DeleteMessage(event.MessageID)
	return nil
}

func TagWithdrawParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	return &tagWithdrawNode{}, nil
}
