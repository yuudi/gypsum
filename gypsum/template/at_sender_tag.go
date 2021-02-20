package template

import (
	"fmt"

	"github.com/flosch/pongo2"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type tagAtSenderNode struct{}

func (node *tagAtSenderNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	eventCtx, ok := ctx.Public["_event"]
	if !ok {
		return ctx.Error("cannot at sender without event", nil)
	}
	event, ok := eventCtx.(*zero.Event)
	if !ok {
		return ctx.Error(fmt.Sprintf("event type error: %#v", eventCtx), nil)
	}
	if event == nil {
		return ctx.Error("event is nil", nil)
	}
	if event.GroupID == 0 {
		return ctx.Error(fmt.Sprintf("cannot at sender in event %s/%s", event.PostType, event.SubType), nil)
	}
	writer.WriteString(fmt.Sprintf("[CQ:at,qq=%d]", event.UserID))
	return nil
}

func TagAtSenderParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	return &tagAtSenderNode{}, nil
}
