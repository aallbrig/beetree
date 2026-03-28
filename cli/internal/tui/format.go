package tui

import (
	"strings"

	"github.com/aallbrig/beetree-cli/internal/model"
)

// nodeTypeTags maps node type strings to short display tags.
var nodeTypeTags = map[string]string{
	"selector":         "[SEL]",
	"sequence":         "[SEQ]",
	"parallel":         "[PAR]",
	"action":           "[ACT]",
	"condition":        "[CND]",
	"decorator":        "[DEC]",
	"utility_selector": "[UTL]",
	"active_selector":  "[ASL]",
	"random_selector":  "[RND]",
	"random_sequence":  "[RSQ]",
	"subtree":          "[SUB]",
}

// NodeTypeTag returns a short bracketed tag for a node type (e.g., "[SEL]").
func NodeTypeTag(nodeType string) string {
	if tag, ok := nodeTypeTags[nodeType]; ok {
		return tag
	}
	return "[???]"
}

// NodeLabel returns a display label for a tree node, including type tag, name,
// and optional suffixes for class references, decorators, or subtree refs.
func NodeLabel(node *model.NodeSpec) string {
	var b strings.Builder
	b.WriteString(NodeTypeTag(node.Type))
	b.WriteString(" ")
	b.WriteString(node.Name)

	if node.Node != "" {
		b.WriteString(" (")
		b.WriteString(node.Node)
		b.WriteString(")")
	}
	if node.Decorator != "" {
		b.WriteString(" ♦")
		b.WriteString(node.Decorator)
	}
	if node.Ref != "" {
		b.WriteString(" →")
		b.WriteString(node.Ref)
	}
	return b.String()
}

// Properties holds computed display data for a node.
type Properties struct {
	Name        string
	Type        string
	ChildCount  int
	NodeClass   string
	Decorator   string
	SubtreeRef  string
	SubtreeFile string
	Parameters  map[string]interface{}
}

// NodeProperties computes display properties for a node.
func NodeProperties(node *model.NodeSpec) Properties {
	if node == nil {
		return Properties{}
	}
	return Properties{
		Name:        node.Name,
		Type:        node.Type,
		ChildCount:  len(node.Children),
		NodeClass:   node.Node,
		Decorator:   node.Decorator,
		SubtreeRef:  node.Ref,
		SubtreeFile: node.File,
		Parameters:  node.Parameters,
	}
}

// NodeTypeEntry represents an available node type for the type selector.
type NodeTypeEntry struct {
	Name     string
	Category string // "CORE", "EXT"
}

// AvailableNodeTypes returns all node types available for selection.
func AvailableNodeTypes() []NodeTypeEntry {
	var types []NodeTypeEntry
	for name := range model.CoreNodeTypes() {
		types = append(types, NodeTypeEntry{Name: name, Category: "CORE"})
	}
	for name := range model.ExtensionNodeTypes() {
		types = append(types, NodeTypeEntry{Name: name, Category: "EXT"})
	}
	return types
}

// FilterNodeTypes returns entries whose name contains the query (case-insensitive).
func FilterNodeTypes(types []NodeTypeEntry, query string) []NodeTypeEntry {
	if query == "" {
		return types
	}
	q := strings.ToLower(query)
	var result []NodeTypeEntry
	for _, nt := range types {
		if strings.Contains(strings.ToLower(nt.Name), q) {
			result = append(result, nt)
		}
	}
	return result
}
