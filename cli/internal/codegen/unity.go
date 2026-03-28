package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/aallbrig/beetree-cli/internal/model"
)

//go:embed templates/unity/*.tmpl
var unityTemplates embed.FS

// UnityGenerator generates Unity C# code from behavior tree specs.
type UnityGenerator struct {
	templates *template.Template
}

// NewUnityGenerator creates a new Unity code generator.
func NewUnityGenerator() *UnityGenerator {
	funcMap := TemplateFuncs()
	funcMap["csharpType"] = csharpType
	funcMap["csharpNodeType"] = csharpNodeType
	funcMap["csharpDefault"] = csharpDefault

	tmpl := template.Must(
		template.New("unity").Funcs(funcMap).ParseFS(unityTemplates, "templates/unity/*.tmpl"),
	)

	return &UnityGenerator{templates: tmpl}
}

func (g *UnityGenerator) Engine() string {
	return "unity"
}

func (g *UnityGenerator) Generate(spec *model.TreeSpec) ([]GeneratedFile, error) {
	data := BuildTemplateData(spec)
	data.SourceFile = spec.Metadata.Name + ".beetree.yaml"

	var files []GeneratedFile

	// Runtime base classes
	rt, err := g.render("runtime.cs.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("runtime: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:    "BTRuntime.cs",
		Content: rt,
		IsStub:  false,
	})

	// Blackboard
	bb, err := g.render("blackboard.cs.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("blackboard: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:    data.TreeClassName + "Blackboard.cs",
		Content: bb,
		IsStub:  false,
	})

	// Tree definition
	td, err := g.render("tree_definition.cs.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("tree definition: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:    data.TreeClassName + "TreeDefinition.cs",
		Content: td,
		IsStub:  false,
	})

	// Custom node stubs (processed first — richer templates with parameters)
	seen := make(map[string]bool)
	for _, cn := range spec.CustomNodes {
		if seen[cn.Name] {
			continue
		}
		seen[cn.Name] = true

		params := make([]ParameterData, len(cn.Parameters))
		for j, p := range cn.Parameters {
			params[j] = ParameterData{Name: p.Name, Type: p.Type, Default: p.Default}
		}

		stubData := struct {
			SourceFile       string
			Name             string
			Description      string
			Parameters       []ParameterData
			BlackboardReads  []string
			BlackboardWrites []string
		}{
			SourceFile:       data.SourceFile,
			Name:             cn.Name,
			Description:      cn.Description,
			Parameters:       params,
			BlackboardReads:  cn.BlackboardReads,
			BlackboardWrites: cn.BlackboardWrites,
		}

		tmplName := "custom_action.cs.tmpl"
		suffix := "Action.cs"
		if cn.Type == "condition" {
			tmplName = "custom_condition.cs.tmpl"
			suffix = "Condition.cs"
		}

		content, err := g.render(tmplName, stubData)
		if err != nil {
			return nil, fmt.Errorf("custom node %s: %w", cn.Name, err)
		}
		files = append(files, GeneratedFile{
			Path:    cn.Name + suffix,
			Content: content,
			IsStub:  true,
		})
	}

	// Action stubs (unique by class name, skip custom nodes already generated)
	actions := CollectActions(&spec.Tree)
	for _, a := range actions {
		className := a.Node
		if className == "" {
			className = ToPascalCase(a.Name)
		}
		if seen[className] {
			continue
		}
		seen[className] = true

		stubData := struct {
			SourceFile string
			ClassName  string
		}{
			SourceFile: data.SourceFile,
			ClassName:  className,
		}

		content, err := g.render("action.cs.tmpl", stubData)
		if err != nil {
			return nil, fmt.Errorf("action %s: %w", className, err)
		}
		files = append(files, GeneratedFile{
			Path:    className + "Action.cs",
			Content: content,
			IsStub:  true,
		})
	}

	// Condition stubs (unique by class name, skip custom nodes already generated)
	conditions := CollectConditions(&spec.Tree)
	for _, c := range conditions {
		className := c.Node
		if className == "" {
			className = ToPascalCase(c.Name)
		}
		if seen[className] {
			continue
		}
		seen[className] = true

		stubData := struct {
			SourceFile string
			ClassName  string
		}{
			SourceFile: data.SourceFile,
			ClassName:  className,
		}

		content, err := g.render("condition.cs.tmpl", stubData)
		if err != nil {
			return nil, fmt.Errorf("condition %s: %w", className, err)
		}
		files = append(files, GeneratedFile{
			Path:    className + "Condition.cs",
			Content: content,
			IsStub:  true,
		})
	}

	return files, nil
}

func (g *UnityGenerator) render(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func csharpType(btType string) string {
	switch strings.ToLower(btType) {
	case "float":
		return "float"
	case "int", "integer":
		return "int"
	case "bool", "boolean":
		return "bool"
	case "string":
		return "string"
	case "vector3":
		return "Vector3"
	case "object":
		return "object"
	default:
		return "object"
	}
}

func csharpDefault(val interface{}) string {
	if val == nil {
		return "default"
	}
	switch v := val.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case float64:
		return fmt.Sprintf("%gf", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func csharpNodeType(btType string) string {
	switch btType {
	case "sequence":
		return "BTSequence"
	case "selector":
		return "BTSelector"
	case "parallel":
		return "BTParallel"
	case "decorator":
		return "BTDecorator"
	case "action":
		return "BTAction"
	case "condition":
		return "BTCondition"
	default:
		return "BTNode"
	}
}
