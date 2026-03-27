package renderer

import (
	"fmt"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/model"
)

var typeAbbreviations = map[string]string{
	"action":           "ACT",
	"condition":        "CND",
	"sequence":         "SEQ",
	"selector":         "SEL",
	"parallel":         "PAR",
	"decorator":        "DEC",
	"utility_selector": "UTL",
	"active_selector":  "ASL",
	"random_selector":  "RSL",
	"random_sequence":  "RSQ",
	"subtree":          "SUB",
}

// RenderSpecASCII renders a NodeSpec as an ASCII tree.
func RenderSpecASCII(node *model.NodeSpec) string {
	var sb strings.Builder
	renderSpecNode(&sb, node, "", true)
	return sb.String()
}

func renderSpecNode(sb *strings.Builder, node *model.NodeSpec, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if prefix == "" {
		connector = ""
	}

	label := nodeLabel(node)
	sb.WriteString(prefix + connector + label + "\n")

	childPrefix := prefix
	if prefix != "" {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for i := range node.Children {
		isLastChild := i == len(node.Children)-1
		renderSpecNode(sb, &node.Children[i], childPrefix, isLastChild)
	}
}

func nodeLabel(node *model.NodeSpec) string {
	abbr := typeAbbreviations[node.Type]
	if abbr == "" {
		abbr = strings.ToUpper(node.Type[:3])
	}

	if node.Type == "decorator" && node.Decorator != "" {
		return fmt.Sprintf("[DEC:%s] %s", node.Decorator, node.Name)
	}

	return fmt.Sprintf("[%s] %s", abbr, node.Name)
}

// RenderMermaid renders a NodeSpec as a Mermaid flowchart diagram.
func RenderMermaid(node *model.NodeSpec) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")
	counter := 0
	renderMermaidNode(&sb, node, "", &counter)
	return sb.String()
}

func renderMermaidNode(sb *strings.Builder, node *model.NodeSpec, parentID string, counter *int) {
	*counter++
	id := fmt.Sprintf("n%d", *counter)

	abbr := typeAbbreviations[node.Type]
	if abbr == "" {
		abbr = strings.ToUpper(node.Type[:3])
	}

	label := fmt.Sprintf("%s: %s", abbr, node.Name)
	if node.Type == "decorator" && node.Decorator != "" {
		label = fmt.Sprintf("DEC:%s: %s", node.Decorator, node.Name)
	}

	shape := nodeShape(node.Type)
	sb.WriteString(fmt.Sprintf("    %s%s\n", id, formatShape(id, label, shape)))

	if parentID != "" {
		sb.WriteString(fmt.Sprintf("    %s --> %s\n", parentID, id))
	}

	for i := range node.Children {
		renderMermaidNode(sb, &node.Children[i], id, counter)
	}
}

func nodeShape(nodeType string) string {
	switch {
	case model.IsLeafType(nodeType):
		return "round"
	case model.IsCompositeType(nodeType):
		return "rect"
	default:
		return "hex"
	}
}

func formatShape(id, label, shape string) string {
	switch shape {
	case "round":
		return fmt.Sprintf("(%s)", label)
	case "hex":
		return fmt.Sprintf("{{%s}}", label)
	default:
		return fmt.Sprintf("[%s]", label)
	}
}
