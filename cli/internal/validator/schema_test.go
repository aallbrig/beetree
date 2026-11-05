package validator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const schemaPath = "../../../website/static/schemas/beetree.schema.json"

func compileSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	raw, err := os.ReadFile(schemaPath)
	require.NoError(t, err)
	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(string(raw)))
	require.NoError(t, err)
	c := jsonschema.NewCompiler()
	require.NoError(t, c.AddResource("https://beetreecraft.com/schemas/beetree.schema.json", doc))
	s, err := c.Compile("https://beetreecraft.com/schemas/beetree.schema.json")
	require.NoError(t, err)
	return s
}

// asJSONValue turns YAML into the plain values a JSON document would decode to.
func asJSONValue(t *testing.T, data []byte) any {
	t.Helper()
	var v any
	require.NoError(t, yaml.Unmarshal(data, &v))
	b, err := json.Marshal(v)
	require.NoError(t, err)
	out, err := jsonschema.UnmarshalJSON(strings.NewReader(string(b)))
	require.NoError(t, err)
	return out
}

// The published schema and the Go validator must agree on every example.
func TestSchemaAcceptsExamples(t *testing.T) {
	s := compileSchema(t)
	files, err := filepath.Glob("../../../examples/*.beetree.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, files)
	for _, f := range files {
		data, err := os.ReadFile(f) //nolint:gosec // repo example files
		require.NoError(t, err)
		assert.NoError(t, s.Validate(asJSONValue(t, data)), f)
		parsed, err := spec.ParseYAML(data)
		require.NoError(t, err)
		assert.Empty(t, Validate(parsed), f)
	}
}

func TestSchemaRejectsWhatTheValidatorRejects(t *testing.T) {
	s := compileSchema(t)
	head := "version: \"1.0\"\nmetadata: {name: t}\n"
	cases := map[string]string{
		"missing metadata name":  "version: \"1.0\"\nmetadata: {}\ntree: {type: action, name: a}\n",
		"missing version":        "metadata: {name: t}\ntree: {type: action, name: a}\n",
		"unknown node type":      head + "tree: {type: wander, name: a}\n",
		"leaf with children":     head + "tree: {type: action, name: a, children: [{type: action, name: b}]}\n",
		"empty composite":        head + "tree: {type: sequence, name: s}\n",
		"decorator two children": head + "tree: {type: decorator, name: d, children: [{type: action, name: a}, {type: action, name: b}]}\n",
		"decorator no child":     head + "tree: {type: decorator, name: d}\n",
		"node without name":      head + "tree: {type: action}\n",
		"bad custom node type":   head + "custom_nodes: [{name: x, type: sequence}]\ntree: {type: action, name: a}\n",
		"subtree without file":   head + "subtrees: [{name: x}]\ntree: {type: action, name: a}\n",
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, s.Validate(asJSONValue(t, []byte(doc))), "schema should reject")
			parsed, err := spec.ParseYAML([]byte(doc))
			if err == nil {
				assert.NotEmpty(t, Validate(parsed), "validator should reject")
			}
		})
	}
}
