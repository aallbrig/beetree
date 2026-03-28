package spec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const minimalYAML = `
version: "1.0"
metadata:
  name: "minimal-tree"
tree:
  type: "selector"
  name: "root"
  children:
    - type: "action"
      name: "do_something"
      node: "DoSomething"
`

const fullYAML = `
version: "1.0"
metadata:
  name: "full-tree"
  description: "A comprehensive test tree"
  author: "tester"
  license: "MIT"
  tags: ["combat", "patrol"]

blackboard:
  - name: "target"
    type: "Entity"
    default: null
    description: "Current target"
  - name: "health"
    type: "float"
    default: 100.0

custom_nodes:
  - name: "DetectEnemy"
    type: "condition"
    description: "Detect nearby enemies"
    category: "perception"
    parameters:
      - name: "radius"
        type: "float"
        default: 15.0
    blackboard_reads: ["position"]
    blackboard_writes: ["target"]

tree:
  type: "selector"
  name: "root"
  children:
    - type: "sequence"
      name: "engage"
      children:
        - type: "condition"
          name: "has_target"
          node: "HasTarget"
        - type: "action"
          name: "attack"
          node: "AttackTarget"
          parameters:
            damage: 10.0
    - type: "action"
      name: "patrol"
      node: "Patrol"

subtrees:
  - name: "flee_subtree"
    file: "./subtrees/flee.beetree.yaml"
`

const decoratorYAML = `
version: "1.0"
metadata:
  name: "decorator-tree"
tree:
  type: "decorator"
  name: "repeat_patrol"
  decorator: "repeat"
  parameters:
    count: 3
  children:
    - type: "action"
      name: "patrol"
      node: "Patrol"
`

const parallelYAML = `
version: "1.0"
metadata:
  name: "parallel-tree"
tree:
  type: "parallel"
  name: "watch_and_move"
  success_policy: "require_all"
  failure_policy: "require_one"
  children:
    - type: "condition"
      name: "is_safe"
      node: "IsSafe"
    - type: "action"
      name: "move"
      node: "Move"
`

func TestParseYAMLString_Minimal(t *testing.T) {
	spec, err := ParseYAML([]byte(minimalYAML))
	require.NoError(t, err)
	require.NotNil(t, spec)

	assert.Equal(t, "1.0", spec.Version)
	assert.Equal(t, "minimal-tree", spec.Metadata.Name)
	assert.Equal(t, "selector", spec.Tree.Type)
	assert.Equal(t, "root", spec.Tree.Name)
	require.Len(t, spec.Tree.Children, 1)
	assert.Equal(t, "action", spec.Tree.Children[0].Type)
	assert.Equal(t, "DoSomething", spec.Tree.Children[0].Node)
}

func TestParseYAMLString_Full(t *testing.T) {
	spec, err := ParseYAML([]byte(fullYAML))
	require.NoError(t, err)

	// Metadata
	assert.Equal(t, "full-tree", spec.Metadata.Name)
	assert.Equal(t, "tester", spec.Metadata.Author)
	assert.Equal(t, "MIT", spec.Metadata.License)
	assert.Equal(t, []string{"combat", "patrol"}, spec.Metadata.Tags)

	// Blackboard
	require.Len(t, spec.Blackboard, 2)
	assert.Equal(t, "target", spec.Blackboard[0].Name)
	assert.Equal(t, "Entity", spec.Blackboard[0].Type)
	assert.Equal(t, "health", spec.Blackboard[1].Name)
	assert.Equal(t, float64(100), spec.Blackboard[1].Default)

	// Custom nodes
	require.Len(t, spec.CustomNodes, 1)
	assert.Equal(t, "DetectEnemy", spec.CustomNodes[0].Name)
	assert.Equal(t, "condition", spec.CustomNodes[0].Type)
	require.Len(t, spec.CustomNodes[0].Parameters, 1)
	assert.Equal(t, "radius", spec.CustomNodes[0].Parameters[0].Name)

	// Tree structure
	assert.Equal(t, "selector", spec.Tree.Type)
	require.Len(t, spec.Tree.Children, 2)

	engage := spec.Tree.Children[0]
	assert.Equal(t, "sequence", engage.Type)
	require.Len(t, engage.Children, 2)
	assert.Equal(t, "condition", engage.Children[0].Type)
	assert.Equal(t, "action", engage.Children[1].Type)
	assert.Equal(t, float64(10), engage.Children[1].Parameters["damage"])

	// Subtrees
	require.Len(t, spec.Subtrees, 1)
	assert.Equal(t, "flee_subtree", spec.Subtrees[0].Name)
}

func TestParseYAMLString_Decorator(t *testing.T) {
	spec, err := ParseYAML([]byte(decoratorYAML))
	require.NoError(t, err)

	assert.Equal(t, "decorator", spec.Tree.Type)
	assert.Equal(t, "repeat", spec.Tree.Decorator)
	assert.Equal(t, 3, spec.Tree.Parameters["count"])
	require.Len(t, spec.Tree.Children, 1)
}

func TestParseYAMLString_Parallel(t *testing.T) {
	spec, err := ParseYAML([]byte(parallelYAML))
	require.NoError(t, err)

	assert.Equal(t, "parallel", spec.Tree.Type)
	assert.Equal(t, "require_all", spec.Tree.SuccessPolicy)
	assert.Equal(t, "require_one", spec.Tree.FailurePolicy)
	require.Len(t, spec.Tree.Children, 2)
}

func TestParseYAMLString_InvalidYAML(t *testing.T) {
	_, err := ParseYAML([]byte("invalid: [yaml: {broken"))
	assert.Error(t, err)
}

func TestParseYAMLString_EmptyInput(t *testing.T) {
	_, err := ParseYAML([]byte(""))
	assert.Error(t, err)
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.beetree.yaml")
	err := os.WriteFile(path, []byte(minimalYAML), 0644)
	require.NoError(t, err)

	spec, err := ParseFile(path)
	require.NoError(t, err)
	assert.Equal(t, "minimal-tree", spec.Metadata.Name)
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/file.beetree.yaml")
	assert.Error(t, err)
}

func TestParseJSON(t *testing.T) {
	jsonData := []byte(`{
		"version": "1.0",
		"metadata": {"name": "json-tree"},
		"tree": {
			"type": "action",
			"name": "do_it",
			"node": "DoIt"
		}
	}`)
	spec, err := ParseJSON(jsonData)
	require.NoError(t, err)
	assert.Equal(t, "json-tree", spec.Metadata.Name)
	assert.Equal(t, "action", spec.Tree.Type)
}

func TestNodeCount(t *testing.T) {
	spec, err := ParseYAML([]byte(fullYAML))
	require.NoError(t, err)

	count := NodeCount(&spec.Tree)
	assert.Equal(t, 5, count) // root(selector) + engage(sequence) + has_target + attack + patrol
}
