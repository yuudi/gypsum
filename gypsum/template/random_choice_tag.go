package template

import (
	"math/rand"

	"github.com/flosch/pongo2"
)

type tagRandomChoiceNode struct {
	wrappers []*pongo2.NodeWrapper
}

func (node *tagRandomChoiceNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	choice := rand.Intn(len(node.wrappers))
	return node.wrappers[choice].Execute(ctx, writer)
}

func TagRandomChoiceParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	node := &tagRandomChoiceNode{}
	for {
		wrapper, _, err := doc.WrapUntilTag("otherwise", "end_random_choice")
		if err != nil {
			return nil, err
		}
		node.wrappers = append(node.wrappers, wrapper)
		if wrapper.Endtag == "end_random_choice" {
			break
		}
	}
	return node, nil
}
