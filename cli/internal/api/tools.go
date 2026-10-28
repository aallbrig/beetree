package api

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/codegen"
	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/renderer"
	"github.com/aallbrig/beetree-cli/internal/simulator"
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/aallbrig/beetree-cli/internal/validator"
)

// Limits that keep stateless tools cheap enough to expose publicly.
const (
	maxSpecBytes = 128 * 1024
	maxNodes     = 500
)

// ValidateResult is the outcome of validating a spec.
type ValidateResult struct {
	Valid     bool     `json:"valid"`
	Errors    []string `json:"errors"`
	NodeCount int      `json:"nodeCount"`
}

// SimStep is one event in a simulation trace.
type SimStep struct {
	Depth  int    `json:"depth"`
	Node   string `json:"node"`
	Type   string `json:"type"`
	Event  string `json:"event"`
	Status string `json:"status,omitempty"`
}

// SimulateResult is a dry-run trace.
type SimulateResult struct {
	Root   string    `json:"root"`
	Status string    `json:"status"`
	Steps  []SimStep `json:"steps"`
	Trace  string    `json:"trace"`
}

// GeneratedFile is one file of generated engine code.
type GeneratedFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	IsStub  bool   `json:"isStub"`
}

var errTooLarge = errors.New("spec too large")

// parseSpec parses and bounds a YAML spec string.
func parseSpec(yamlText string) (*model.TreeSpec, error) {
	if len(yamlText) > maxSpecBytes {
		return nil, errTooLarge
	}
	s, err := spec.ParseYAML([]byte(yamlText))
	if err != nil {
		return nil, err
	}
	if n := spec.NodeCount(&s.Tree); n > maxNodes {
		return nil, fmt.Errorf("tree has %d nodes; limit is %d", n, maxNodes)
	}
	return s, nil
}

// validateYAML parses and validates; parse failures are reported as errors in
// the result rather than as a Go error.
func validateYAML(yamlText string) (ValidateResult, *model.TreeSpec) {
	s, err := parseSpec(yamlText)
	if err != nil {
		return ValidateResult{Errors: []string{err.Error()}}, nil
	}
	res := ValidateResult{NodeCount: spec.NodeCount(&s.Tree), Errors: []string{}}
	for _, e := range validator.Validate(s) {
		res.Errors = append(res.Errors, e.Error())
	}
	res.Valid = len(res.Errors) == 0
	return res, s
}

// validSpec returns the parsed spec or an error describing why it is invalid.
func validSpec(yamlText string) (*model.TreeSpec, error) {
	res, s := validateYAML(yamlText)
	if !res.Valid {
		return nil, errors.New(strings.Join(res.Errors, "; "))
	}
	return s, nil
}

func simulateYAML(yamlText string, overrides map[string]string) (*SimulateResult, error) {
	s, err := validSpec(yamlText)
	if err != nil {
		return nil, err
	}
	ov := map[string]simulator.Status{}
	for name, st := range overrides {
		switch simulator.Status(strings.ToUpper(st)) {
		case simulator.StatusSuccess, simulator.StatusFailure, simulator.StatusRunning:
			ov[name] = simulator.Status(strings.ToUpper(st))
		default:
			return nil, fmt.Errorf("override %q: status must be SUCCESS, FAILURE or RUNNING", name)
		}
	}
	r := simulator.Simulate(s, ov)
	out := &SimulateResult{Root: r.NodeName, Status: string(r.Status), Trace: simulator.FormatTrace(r), Steps: []SimStep{}}
	for _, st := range r.Steps {
		out.Steps = append(out.Steps, SimStep{Depth: st.Depth, Node: st.NodeName, Type: st.NodeType, Event: st.Event, Status: string(st.Status)})
	}
	return out, nil
}

// RenderFormats lists supported render formats.
var RenderFormats = []string{"ascii", "sigil", "mermaid", "dot"}

func renderYAML(yamlText, format string) (string, error) {
	s, err := validSpec(yamlText)
	if err != nil {
		return "", err
	}
	switch format {
	case "", "ascii":
		return renderer.RenderSpecASCII(&s.Tree), nil
	case "sigil":
		return renderer.RenderSigil(&s.Tree, s.Notation), nil
	case "mermaid":
		return renderer.RenderMermaid(&s.Tree), nil
	case "dot":
		return renderer.RenderDOT(&s.Tree), nil
	}
	return "", fmt.Errorf("unknown format %q (want one of %s)", format, strings.Join(RenderFormats, ", "))
}

// Engines lists supported code generation targets.
var Engines = []string{"godot", "unity", "unreal"}

func generateYAML(yamlText, engine string) ([]GeneratedFile, error) {
	s, err := validSpec(yamlText)
	if err != nil {
		return nil, err
	}
	var g codegen.Generator
	switch engine {
	case "unity":
		g = codegen.NewUnityGenerator()
	case "unreal":
		g = codegen.NewUnrealGenerator()
	case "godot":
		g = codegen.NewGodotGenerator()
	default:
		return nil, fmt.Errorf("unknown engine %q (want one of %s)", engine, strings.Join(Engines, ", "))
	}
	files, err := g.Generate(s)
	if err != nil {
		return nil, err
	}
	out := make([]GeneratedFile, 0, len(files))
	for _, f := range files {
		out = append(out, GeneratedFile{Path: f.Path, Content: f.Content, IsStub: f.IsStub})
	}
	return out, nil
}

// NodeTypeInfo describes a node type for clients building editors.
type NodeTypeInfo struct {
	Type      string `json:"type"`
	Composite bool   `json:"composite"`
	Leaf      bool   `json:"leaf"`
}

func nodeTypes() ([]NodeTypeInfo, []string) {
	var types []NodeTypeInfo
	for t := range model.CoreNodeTypes() {
		types = append(types, NodeTypeInfo{Type: t, Composite: model.IsCompositeType(t), Leaf: model.IsLeafType(t)})
	}
	sort.Slice(types, func(i, j int) bool { return types[i].Type < types[j].Type })
	var decs []string
	for d := range model.BuiltinDecorators() {
		decs = append(decs, d)
	}
	sort.Strings(decs)
	return types, decs
}
