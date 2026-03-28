package renderer

import (
	"fmt"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/model"
)

// Built-in Unicode sigils for node types, based on the BeeTree Sigil System (BSS).
// Sequence (→) and Selector (?) follow academic BT convention (Colledanchise & Ögren).
var defaultTypeSigils = map[string]string{
	"action":           "!",
	"condition":        "¿",
	"sequence":         "→",
	"selector":         "?",
	"parallel":         "⇒",
	"decorator":        "◇",
	"utility_selector": "⚖",
	"active_selector":  "⚡",
	"random_selector":  "?~",
	"random_sequence":  "→~",
	"subtree":          "↗",
}

// decoratorSigils maps specific decorator names to sub-sigils.
var decoratorSigils = map[string]string{
	"repeat":         "◇∞",
	"negate":         "◇¬",
	"always_succeed": "◇✓",
	"always_fail":    "◇✗",
	"until_fail":     "◇⇣",
	"until_succeed":  "◇⇡",
	"timeout":        "◇⏱",
	"cooldown":       "◇⏳",
	"retry":          "◇↻",
}

// ResolveSigil returns the sigil for a node, respecting custom overrides.
// Resolution order: node-class sigil > type sigil override > decorator sub-sigil > default.
func ResolveSigil(node *model.NodeSpec, notation model.NotationConfig) string {
	// 1. Check node-class sigil (by Node field, e.g. "PlayAnimation")
	if node.Node != "" && notation.NodeSigils != nil {
		if sigil, ok := notation.NodeSigils[node.Node]; ok {
			return sigil
		}
	}

	// 2. Check type sigil override
	if notation.TypeSigils != nil {
		if sigil, ok := notation.TypeSigils[node.Type]; ok {
			if node.Type == "decorator" && node.Decorator != "" {
				return sigil + ":" + node.Decorator
			}
			return sigil
		}
	}

	// 3. Check decorator sub-sigil
	if node.Type == "decorator" && node.Decorator != "" {
		if sigil, ok := decoratorSigils[node.Decorator]; ok {
			return sigil
		}
		return "◇" + node.Decorator
	}

	// 4. Fall back to default sigil table
	if sigil, ok := defaultTypeSigils[node.Type]; ok {
		return sigil
	}

	// Unknown type: use first 3 chars uppercased
	if len(node.Type) >= 3 {
		return strings.ToUpper(node.Type[:3])
	}
	return strings.ToUpper(node.Type)
}

// RenderSigil renders a NodeSpec as a tree with Unicode sigils and box-drawing characters.
func RenderSigil(node *model.NodeSpec, notation model.NotationConfig) string {
	var sb strings.Builder
	renderSigilNode(&sb, node, notation, "", true, true)
	return sb.String()
}

func renderSigilNode(sb *strings.Builder, node *model.NodeSpec, notation model.NotationConfig, prefix string, isLast bool, isRoot bool) {
	sigil := ResolveSigil(node, notation)
	label := fmt.Sprintf("%s %s", sigil, node.Name)

	if isRoot {
		sb.WriteString(label + "\n")
	} else {
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		sb.WriteString(prefix + connector + label + "\n")
	}

	var childPrefix string
	if isRoot {
		childPrefix = ""
	} else if isLast {
		childPrefix = prefix + "    "
	} else {
		childPrefix = prefix + "│   "
	}

	for i := range node.Children {
		isLastChild := i == len(node.Children)-1
		renderSigilNode(sb, &node.Children[i], notation, childPrefix, isLastChild, false)
	}
}

// RenderCompact renders a NodeSpec using sigils with indentation only (no box-drawing).
// This format is ideal for diffing.
func RenderCompact(node *model.NodeSpec, notation model.NotationConfig) string {
	var sb strings.Builder
	renderCompactNode(&sb, node, notation, 0)
	return sb.String()
}

func renderCompactNode(sb *strings.Builder, node *model.NodeSpec, notation model.NotationConfig, depth int) {
	sigil := ResolveSigil(node, notation)
	indent := strings.Repeat("  ", depth)
	sb.WriteString(fmt.Sprintf("%s%s %s\n", indent, sigil, node.Name))

	for i := range node.Children {
		renderCompactNode(sb, &node.Children[i], notation, depth+1)
	}
}

// RenderOneline renders a NodeSpec as a single-line S-expression.
// Composite/decorator nodes: sigil name(child child ...). Leaf nodes: sigil name.
func RenderOneline(node *model.NodeSpec, notation model.NotationConfig) string {
	var sb strings.Builder
	renderOnelineNode(&sb, node, notation)
	return sb.String()
}

func renderOnelineNode(sb *strings.Builder, node *model.NodeSpec, notation model.NotationConfig) {
	sigil := ResolveSigil(node, notation)

	if len(node.Children) == 0 {
		sb.WriteString(sigil)
		sb.WriteString(node.Name)
		return
	}

	sb.WriteString(sigil)
	sb.WriteString(node.Name)
	sb.WriteString("(")
	for i := range node.Children {
		if i > 0 {
			sb.WriteString(" ")
		}
		renderOnelineNode(sb, &node.Children[i], notation)
	}
	sb.WriteString(")")
}
