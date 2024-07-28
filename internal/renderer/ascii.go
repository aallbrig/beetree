package renderer

import (
	"github.com/aallbrig/beetree-cli/internal/parser"
)

func RenderASCII(node *parser.Node) (string, error) {
	return renderNode(node, "", true), nil
}

func renderNode(node *parser.Node, prefix string, isLast bool) string {
	result := prefix

	if len(prefix) > 0 {
		if isLast {
			result += "└── "
		} else {
			result += "├── "
		}
	}

	result += node.Type + "\n"

	for i, child := range node.Children {
		newPrefix := prefix
		if len(prefix) > 0 {
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
		}
		result += renderNode(child, newPrefix, i == len(node.Children)-1)
	}

	return result
}
