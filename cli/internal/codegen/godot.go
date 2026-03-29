package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/aallbrig/beetree-cli/internal/model"
)

//go:embed templates/godot/*.tmpl
var godotTemplates embed.FS

// GodotGenerator generates Godot GDScript code from behavior tree specs.
type GodotGenerator struct {
	templates *template.Template
}

// NewGodotGenerator creates a new Godot code generator.
func NewGodotGenerator() *GodotGenerator {
	funcMap := TemplateFuncs()
	funcMap["gdscriptType"] = gdscriptType
	funcMap["gdscriptDefault"] = gdscriptDefault

	tmpl := template.Must(
		template.New("godot").Funcs(funcMap).ParseFS(godotTemplates, "templates/godot/*.tmpl"),
	)

	return &GodotGenerator{templates: tmpl}
}

func (g *GodotGenerator) Engine() string {
	return "godot"
}

func (g *GodotGenerator) Generate(spec *model.TreeSpec) ([]GeneratedFile, error) {
	data := BuildTemplateData(spec)
	data.SourceFile = spec.Metadata.Name + ".beetree.yaml"

	snakeName := ToSnakeCase(data.TreeClassName)
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
	rt, err := g.render("runtime.gd.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("runtime: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:    "bt_runtime.gd",
		Content: rt,
		IsStub:  false,
	})

	// Blackboard
	bb, err := g.render("blackboard.gd.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("blackboard: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:    snakeName + "_blackboard.gd",
		Content: bb,
		IsStub:  false,
	})

	// Tree definition
	td, err := g.render("tree_definition.gd.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("tree definition: %w", err)
	}
	files = append(files, GeneratedFile{
		Path:    snakeName + "_tree_definition.gd",
		Content: td,
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

		tmplName := "custom_action.gd.tmpl"
		suffix := "_action.gd"
		if cn.Type == "condition" {
			tmplName = "custom_condition.gd.tmpl"
			suffix = "_condition.gd"
		}

		content, err := g.render(tmplName, stubData)
		if err != nil {
			return nil, fmt.Errorf("custom node %s: %w", cn.Name, err)
		}
		files = append(files, GeneratedFile{
			Path:    ToSnakeCase(cn.Name) + suffix,
			Content: content,
			IsStub:  true,
		})
	}

	// Action stubs
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

		stubData := StubData{
			SourceFile:  data.SourceFile,
			ClassName:   className,
			Description: a.Description,
			Parameters:  a.Parameters,
			Blackboard:  data.Blackboard,
		}

		content, err := g.render("action.gd.tmpl", stubData)
		if err != nil {
			return nil, fmt.Errorf("action %s: %w", className, err)
		}
		files = append(files, GeneratedFile{
			Path:    ToSnakeCase(className) + "_action.gd",
			Content: content,
			IsStub:  true,
		})
	}

	// Condition stubs
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

		stubData := StubData{
			SourceFile:  data.SourceFile,
			ClassName:   className,
			Description: c.Description,
			Parameters:  c.Parameters,
			Blackboard:  data.Blackboard,
		}

		content, err := g.render("condition.gd.tmpl", stubData)
		if err != nil {
			return nil, fmt.Errorf("condition %s: %w", className, err)
		}
		files = append(files, GeneratedFile{
			Path:    ToSnakeCase(className) + "_condition.gd",
			Content: content,
			IsStub:  true,
		})
	}

	return files, nil
}

func (g *GodotGenerator) render(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func gdscriptType(btType string) string {
	switch strings.ToLower(btType) {
	case "float":
		return "float"
	case "int", "integer":
		return "int"
	case "bool", "boolean":
		return "bool"
	case "string":
		return "String"
	case "vector3":
		return "Vector3"
	case "object":
		return "Variant"
	default:
		return "Variant"
	}
}

func gdscriptDefault(val interface{}) string {
	if val == nil {
		return "null"
	}
	switch v := val.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		return fmt.Sprintf("\"%s\"", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
