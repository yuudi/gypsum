package template

import (
	"strconv"
	"strings"

	"github.com/flosch/pongo2"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type MessageType int

const (
	PrivateMessageType MessageType = iota
	GroupMessageType
)

type tagSendNode struct {
	targetType MessageType
	targetID   int64
	wrapper    *pongo2.NodeWrapper
}

func (node *tagSendNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	var message = &strings.Builder{}
	if err := node.wrapper.Execute(ctx, message); err != nil {
		return err
	}
	messageSend := strings.TrimSpace(message.String())
	if len(messageSend) == 0 {
		return nil
	}
	switch node.targetType {
	case PrivateMessageType:
		zero.SendPrivateMessage(node.targetID, messageSend)
	case GroupMessageType:
		zero.SendGroupMessage(node.targetID, messageSend)
	}
	return nil
}

func TagSendParser(msgType MessageType) pongo2.TagParser {
	return func(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
		sendNode := &tagSendNode{targetType: msgType}
		wrapper, _, pErr := doc.WrapUntilTag("endsend", "end_send")
		if pErr != nil {
			return nil, pErr
		}
		targetID := arguments.MatchType(pongo2.TokenNumber)
		if targetID == nil {
			return nil, arguments.Error("a number token is required", nil)
		}
		sendNode.wrapper = wrapper
		var err error
		sendNode.targetID, err = strconv.ParseInt(targetID.Val, 10, 64)
		if err != nil {
			return nil, arguments.Error("a number token is required", nil)
		}
		return sendNode, nil
	}
}
