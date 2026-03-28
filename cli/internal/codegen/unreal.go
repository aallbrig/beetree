package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/aallbrig/beetree-cli/internal/model"
)

//go:embed templates/unreal/*.tmpl
var unrealTemplates embed.FS

// UnrealGenerator generates Unreal Engine C++ code from behavior tree specs.
type UnrealGenerator struct {
	templates *template.Template
}

// NewUnrealGenerator creates a new Unreal code generator.
func NewUnrealGenerator() *UnrealGenerator {
	funcMap := TemplateFuncs()
	funcMap["unrealType"] = unrealType

	tmpl := template.Must(
		template.New("unreal").Funcs(funcMap).ParseFS(unrealTemplates, "templates/unreal/*.tmpl"),
	)

	return &UnrealGenerator{templates: tmpl}
}

func (g *UnrealGenerator) Engine() string {
	return "unreal"
}

func (g *UnrealGenerator) Generate(spec *model.TreeSpec) ([]GeneratedFile, error) {
	data := BuildTemplateData(spec)
	data.SourceFile = spec.Metadata.Name + ".beetree.yaml"

	var files []GeneratedFile

	// Integration guide README
	readme, err := g.render("readme.md.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("readme: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:    "README.md",
		Content: readme,
		IsStub:  false,
	})

	// Runtime base classes
	rt, err := g.render("runtime.h.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("runtime: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:    "BTRuntime.h",
		Content: rt,
		IsStub:  false,
	})

	// Blackboard header
	bb, err := g.render("blackboard.h.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("blackboard: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:    data.TreeClassName + "Blackboard.h",
		Content: bb,
		IsStub:  false,
	})

	// Tree definition header + source
	tdH, err := g.render("tree_definition.h.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("tree definition header: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:    data.TreeClassName + "TreeDefinition.h",
		Content: tdH,
		IsStub:  false,
	})

	tdCpp, err := g.render("tree_definition.cpp.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("tree definition source: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:    data.TreeClassName + "TreeDefinition.cpp",
		Content: tdCpp,
		IsStub:  false,
	})

	// Custom node stubs (processed first for richer templates)
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
		}{data.SourceFile, cn.Name, cn.Description, params, cn.BlackboardReads, cn.BlackboardWrites}

		if cn.Type == "action" {
			h, err := g.render("custom_task.h.tmpl", stubData)
			if err != nil {
				return nil, fmt.Errorf("custom task header %s: %w", cn.Name, err)
			}
			cpp, err := g.render("custom_task.cpp.tmpl", stubData)
			if err != nil {
				return nil, fmt.Errorf("custom task source %s: %w", cn.Name, err)
			}
			files = append(files, GeneratedFile{Path: "BTTask_" + cn.Name + ".h", Content: h, IsStub: true})
			files = append(files, GeneratedFile{Path: "BTTask_" + cn.Name + ".cpp", Content: cpp, IsStub: true})
		} else {
			h, err := g.render("custom_decorator.h.tmpl", stubData)
			if err != nil {
				return nil, fmt.Errorf("custom decorator header %s: %w", cn.Name, err)
			}
			cpp, err := g.render("custom_decorator.cpp.tmpl", stubData)
			if err != nil {
				return nil, fmt.Errorf("custom decorator source %s: %w", cn.Name, err)
			}
			files = append(files, GeneratedFile{Path: "BTDecorator_" + cn.Name + ".h", Content: h, IsStub: true})
			files = append(files, GeneratedFile{Path: "BTDecorator_" + cn.Name + ".cpp", Content: cpp, IsStub: true})
		}
	}

	// Task stubs for actions
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
		}{data.SourceFile, className}

		h, err := g.render("task.h.tmpl", stubData)
		if err != nil {
			return nil, fmt.Errorf("task header %s: %w", className, err)
		}
		cpp, err := g.render("task.cpp.tmpl", stubData)
		if err != nil {
			return nil, fmt.Errorf("task source %s: %w", className, err)
		}

		files = append(files, GeneratedFile{Path: "BTTask_" + className + ".h", Content: h, IsStub: true})
		files = append(files, GeneratedFile{Path: "BTTask_" + className + ".cpp", Content: cpp, IsStub: true})
	}

	// Decorator stubs for conditions
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
		}{data.SourceFile, className}

		h, err := g.render("decorator.h.tmpl", stubData)
		if err != nil {
			return nil, fmt.Errorf("decorator header %s: %w", className, err)
		}
		cpp, err := g.render("decorator.cpp.tmpl", stubData)
		if err != nil {
			return nil, fmt.Errorf("decorator source %s: %w", className, err)
		}

		files = append(files, GeneratedFile{Path: "BTDecorator_" + className + ".h", Content: h, IsStub: true})
		files = append(files, GeneratedFile{Path: "BTDecorator_" + className + ".cpp", Content: cpp, IsStub: true})
	}

	return files, nil
}

func (g *UnrealGenerator) render(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func unrealType(btType string) string {
	switch strings.ToLower(btType) {
	case "float":
		return "float"
	case "int", "integer":
		return "int32"
	case "bool", "boolean":
		return "bool"
	case "string":
		return "FString"
	case "vector3":
		return "FVector"
	case "object":
		return "UObject*"
	default:
		return "UObject*"
	}
}
