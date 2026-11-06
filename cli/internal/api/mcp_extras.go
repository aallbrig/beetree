package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

// starterSpec is a small valid tree served as an MCP resource so a model has a
// format example before it authors one.
const starterSpec = `version: "1.0"
metadata:
  name: patrol
  description: "NPC patrols between waypoints in a loop"
tree:
  type: decorator
  name: patrol_loop
  decorator: repeat
  children:
    - type: sequence
      name: patrol_cycle
      children:
        - type: condition
          name: has_waypoint
          node: HasWaypoint
        - type: action
          name: move_to_waypoint
          node: MoveToWaypoint
        - type: action
          name: wait_at_waypoint
          node: WaitAtWaypoint
`

// schemaJSON is the tree format JSON Schema. It is a copy of
// website/static/schemas/beetree.schema.json (go:embed cannot reach outside
// the module); TestMCPSchemaMatchesPublished keeps the two identical.
//
//go:embed beetree.schema.json
var schemaJSON string

// templatesFS holds the builder templates (web/src/samples.ts) as .beetree.yaml
// files; TestMCPTemplatesMatchBuilder keeps the two in step.
//
//go:embed templates/*.beetree.yaml
var templatesFS embed.FS

const templateURIPrefix = "beetree://templates/"

type mcpResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

func mcpResources() []mcpResource {
	list := []mcpResource{
		{URI: "beetree://node-types", Name: "Node types", Description: "Node types, decorators, render formats and engines as JSON.", MimeType: "application/json"},
		{URI: "beetree://schema/beetree", Name: "Tree format JSON Schema", Description: "The JSON Schema for .beetree.yaml, as published at /schemas/beetree.schema.json.", MimeType: "application/schema+json"},
		{URI: "beetree://examples/patrol", Name: "Starter spec: patrol", Description: "A small valid .beetree.yaml to copy the format from.", MimeType: "application/yaml"},
	}
	entries, _ := templatesFS.ReadDir("templates")
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".beetree.yaml")
		list = append(list, mcpResource{URI: templateURIPrefix + name, Name: "Template: " + name, Description: "A builder template as a valid .beetree.yaml to start from.", MimeType: "application/yaml"})
	}
	return list
}

func readMCPResource(uri string) (string, string, bool) {
	switch uri {
	case "beetree://node-types":
		types, decs := nodeTypes()
		b, _ := json.Marshal(map[string]any{"nodeTypes": types, "decorators": decs, "engines": Engines, "renderFormats": RenderFormats})
		return string(b), "application/json", true
	case "beetree://schema/beetree":
		return schemaJSON, "application/schema+json", true
	case "beetree://examples/patrol":
		return starterSpec, "application/yaml", true
	}
	if name, ok := strings.CutPrefix(uri, templateURIPrefix); ok && !strings.ContainsAny(name, "/.\\") {
		if b, err := templatesFS.ReadFile("templates/" + name + ".beetree.yaml"); err == nil {
			return string(b), "application/yaml", true
		}
	}
	return "", "", false
}

type mcpPromptArg struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type mcpPrompt struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Arguments   []mcpPromptArg `json:"arguments"`
}

func mcpPrompts() []mcpPrompt {
	return []mcpPrompt{{
		Name:        "design_behavior_tree",
		Description: "Design a behavior tree for a game character, then validate, simulate and generate code for it.",
		Arguments: []mcpPromptArg{
			{Name: "character", Description: "who the tree controls and what they should do", Required: true},
			{Name: "engine", Description: "unity | unreal | godot (optional)", Required: false},
		},
	}}
}

func promptText(name string, args map[string]string) (string, error) {
	if name != "design_behavior_tree" {
		return "", fmt.Errorf("unknown prompt: %s", name)
	}
	character := strings.TrimSpace(args["character"])
	if character == "" {
		return "", fmt.Errorf("prompt design_behavior_tree needs a character argument")
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Design a behavior tree for this game character: %s\n\n", character)
	sb.WriteString("1. Read the beetree://schema/beetree, beetree://node-types and beetree://examples/patrol resources for the format.\n")
	sb.WriteString("2. Write the spec as .beetree.yaml text.\n")
	sb.WriteString("3. Call validate_tree and fix every error it reports.\n")
	sb.WriteString("4. Call simulate_tree (or simulate_start and simulate_answer) to check the failure paths behave as intended.\n")
	sb.WriteString("5. Call render_tree with format ascii and show the result.\n")
	if e := strings.TrimSpace(args["engine"]); e != "" {
		fmt.Fprintf(&sb, "6. Call generate_code with engine %s.\n", e)
	}
	return sb.String(), nil
}
