package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"unicode"

	"github.com/aallbrig/beetree-cli/internal/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GeneratedFile represents a single file output by code generation.
type GeneratedFile struct {
	Path     string
	Content  string
	IsStub   bool // true = user-editable stub; false = regenerated each run
}

// Generator is the interface for engine-specific code generators.
type Generator interface {
	Engine() string
	Generate(spec *model.TreeSpec) ([]GeneratedFile, error)
}

// TemplateData holds all data passed to code generation templates.
type TemplateData struct {
	TreeName      string
	TreeClassName string
	Description   string
	SourceFile    string
	Blackboard    []BlackboardVarData
	Actions       []NodeData
	Conditions    []NodeData
	CustomNodes   []CustomNodeData
	RootNode      *NodeData
	AllNodes      []NodeData
}

// BlackboardVarData is template-friendly blackboard variable info.
type BlackboardVarData struct {
	Name        string
	Type        string
	Default     interface{}
	Description string
}

// NodeData is template-friendly node info.
type NodeData struct {
	Name       string
	ClassName  string
	Type       string
	Parameters map[string]interface{}
	Children   []NodeData
	Decorator  string
}

// CustomNodeData is template-friendly custom node definition.
type CustomNodeData struct {
	Name             string
	ClassName        string
	Type             string
	Description      string
	Parameters       []ParameterData
	BlackboardReads  []string
	BlackboardWrites []string
}

// ParameterData is template-friendly parameter info.
type ParameterData struct {
	Name    string
	Type    string
	Default interface{}
}

// BuildTemplateData converts a TreeSpec into TemplateData for templates.
func BuildTemplateData(spec *model.TreeSpec) *TemplateData {
	rootData := nodeToData(&spec.Tree)
	actions := CollectActions(&spec.Tree)
	conditions := CollectConditions(&spec.Tree)

	bbVars := make([]BlackboardVarData, len(spec.Blackboard))
	for i, bv := range spec.Blackboard {
		bbVars[i] = BlackboardVarData{
			Name:        bv.Name,
			Type:        bv.Type,
			Default:     bv.Default,
			Description: bv.Description,
		}
	}

	customNodes := make([]CustomNodeData, len(spec.CustomNodes))
	for i, cn := range spec.CustomNodes {
		params := make([]ParameterData, len(cn.Parameters))
		for j, p := range cn.Parameters {
			params[j] = ParameterData{Name: p.Name, Type: p.Type, Default: p.Default}
		}
		customNodes[i] = CustomNodeData{
			Name:             cn.Name,
			ClassName:        cn.Name,
			Type:             cn.Type,
			Description:      cn.Description,
			Parameters:       params,
			BlackboardReads:  cn.BlackboardReads,
			BlackboardWrites: cn.BlackboardWrites,
		}
	}

	actionData := make([]NodeData, len(actions))
	for i, a := range actions {
		actionData[i] = nodeToData(a)
	}

	condData := make([]NodeData, len(conditions))
	for i, c := range conditions {
		condData[i] = nodeToData(c)
	}

	return &TemplateData{
		TreeName:      spec.Metadata.Name,
		TreeClassName: ToPascalCase(spec.Metadata.Name),
		Description:   spec.Metadata.Description,
		Blackboard:    bbVars,
		Actions:       actionData,
		Conditions:    condData,
		CustomNodes:   customNodes,
		RootNode:      &rootData,
		AllNodes:      collectAllNodeData(&spec.Tree),
	}
}

func nodeToData(n *model.NodeSpec) NodeData {
	className := n.Node
	if className == "" {
		className = ToPascalCase(n.Name)
	}

	children := make([]NodeData, len(n.Children))
	for i := range n.Children {
		children[i] = nodeToData(&n.Children[i])
	}

	return NodeData{
		Name:       n.Name,
		ClassName:  className,
		Type:       n.Type,
		Parameters: n.Parameters,
		Children:   children,
		Decorator:  n.Decorator,
	}
}

func collectAllNodeData(node *model.NodeSpec) []NodeData {
	var result []NodeData
	result = append(result, nodeToData(node))
	for i := range node.Children {
		result = append(result, collectAllNodeData(&node.Children[i])...)
	}
	return result
}

// CollectLeafNodes returns all leaf nodes (action/condition) from the tree.
func CollectLeafNodes(node *model.NodeSpec) []*model.NodeSpec {
	var leaves []*model.NodeSpec
	collectLeaves(node, &leaves)
	return leaves
}

func collectLeaves(node *model.NodeSpec, out *[]*model.NodeSpec) {
	if model.IsLeafType(node.Type) {
		*out = append(*out, node)
	}
	for i := range node.Children {
		collectLeaves(&node.Children[i], out)
	}
}

// CollectActions returns all action nodes from the tree.
func CollectActions(node *model.NodeSpec) []*model.NodeSpec {
	var result []*model.NodeSpec
	collectByType(node, "action", &result)
	return result
}

// CollectConditions returns all condition nodes from the tree.
func CollectConditions(node *model.NodeSpec) []*model.NodeSpec {
	var result []*model.NodeSpec
	collectByType(node, "condition", &result)
	return result
}

func collectByType(node *model.NodeSpec, nodeType string, out *[]*model.NodeSpec) {
	if node.Type == nodeType {
		*out = append(*out, node)
	}
	for i := range node.Children {
		collectByType(&node.Children[i], nodeType, out)
	}
}

// CollectUniqueNodeClasses returns unique Node class names from the tree.
func CollectUniqueNodeClasses(node *model.NodeSpec) []string {
	seen := make(map[string]bool)
	var result []string
	collectClasses(node, seen, &result)
	return result
}

func collectClasses(node *model.NodeSpec, seen map[string]bool, out *[]string) {
	if node.Node != "" && !seen[node.Node] {
		seen[node.Node] = true
		*out = append(*out, node.Node)
	}
	for i := range node.Children {
		collectClasses(&node.Children[i], seen, out)
	}
}

// ToPascalCase converts kebab-case or snake_case to PascalCase.
func ToPascalCase(s string) string {
	var result strings.Builder
	capitalizeNext := true
	for _, r := range s {
		if r == '-' || r == '_' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(unicode.ToLower(r))
		}
	}
	return result.String()
}

// ToSnakeCase converts PascalCase/camelCase to snake_case.
func ToSnakeCase(s string) string {
	var result strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				// Only insert underscore if previous char is lowercase
				// or next char is lowercase (handles "BTNode" → "bt_node")
				prev := runes[i-1]
				if unicode.IsLower(prev) {
					result.WriteRune('_')
				} else if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// GeneratedFileHeader returns the auto-generated header for files that get regenerated.
func GeneratedFileHeader(sourceFile, commentPrefix string) string {
	return fmt.Sprintf(`%s ============================================================
%s AUTO-GENERATED by BeeTree CLI — DO NOT EDIT
%s Source: %s
%s Regenerated on each beetree generate run
%s ============================================================`,
		commentPrefix, commentPrefix, commentPrefix, sourceFile, commentPrefix, commentPrefix)
}

// StubFileHeader returns the header for user-editable stub files.
func StubFileHeader(sourceFile, commentPrefix string) string {
	return fmt.Sprintf(`%s ============================================================
%s Generated by BeeTree CLI — EDIT THIS FILE
%s Source: %s
%s Implement your custom logic below
%s ============================================================`,
		commentPrefix, commentPrefix, commentPrefix, sourceFile, commentPrefix, commentPrefix)
}

// TemplateFuncs returns common template functions available in all templates.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"pascal":    ToPascalCase,
		"snake":     ToSnakeCase,
		"lower":     strings.ToLower,
		"upper":     strings.ToUpper,
		"title":     cases.Title(language.English).String,
		"genHeader": GeneratedFileHeader,
		"stubHeader": StubFileHeader,
	}
}

// GenerateFromTemplate renders a single named template with the given data.
func GenerateFromTemplate(name, tmplStr string, data *TemplateData) (string, error) {
	tmpl, err := template.New(name).Funcs(TemplateFuncs()).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", name, err)
	}

	return buf.String(), nil
}
