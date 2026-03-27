package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeStatus(t *testing.T) {
	assert.Equal(t, "Success", StatusSuccess.String())
	assert.Equal(t, "Failure", StatusFailure.String())
	assert.Equal(t, "Running", StatusRunning.String())
}

func TestCoreNodeTypes(t *testing.T) {
	coreTypes := CoreNodeTypes()
	require.Len(t, coreTypes, 6)

	expected := []string{"action", "condition", "sequence", "selector", "parallel", "decorator"}
	for _, name := range expected {
		assert.Contains(t, coreTypes, name, "core types should contain %s", name)
	}
}

func TestIsValidNodeType(t *testing.T) {
	assert.True(t, IsValidNodeType("action"))
	assert.True(t, IsValidNodeType("condition"))
	assert.True(t, IsValidNodeType("sequence"))
	assert.True(t, IsValidNodeType("selector"))
	assert.True(t, IsValidNodeType("parallel"))
	assert.True(t, IsValidNodeType("decorator"))
	assert.False(t, IsValidNodeType("invalid"))
	assert.False(t, IsValidNodeType(""))
}

func TestIsCompositeType(t *testing.T) {
	assert.True(t, IsCompositeType("sequence"))
	assert.True(t, IsCompositeType("selector"))
	assert.True(t, IsCompositeType("parallel"))
	assert.False(t, IsCompositeType("action"))
	assert.False(t, IsCompositeType("condition"))
	assert.False(t, IsCompositeType("decorator"))
}

func TestIsLeafType(t *testing.T) {
	assert.True(t, IsLeafType("action"))
	assert.True(t, IsLeafType("condition"))
	assert.False(t, IsLeafType("sequence"))
	assert.False(t, IsLeafType("decorator"))
}

func TestTreeSpec(t *testing.T) {
	spec := &TreeSpec{
		Version: "1.0",
		Metadata: Metadata{
			Name:        "test-tree",
			Description: "A test behavior tree",
			Author:      "tester",
			Tags:        []string{"test"},
		},
		Blackboard: []BlackboardVar{
			{Name: "health", Type: "float", Default: 100.0, Description: "Character health"},
		},
		Tree: NodeSpec{
			Type: "selector",
			Name: "root",
			Children: []NodeSpec{
				{
					Type: "action",
					Name: "attack",
					Node: "AttackTarget",
				},
			},
		},
	}

	assert.Equal(t, "1.0", spec.Version)
	assert.Equal(t, "test-tree", spec.Metadata.Name)
	require.Len(t, spec.Blackboard, 1)
	assert.Equal(t, "health", spec.Blackboard[0].Name)
	assert.Equal(t, "selector", spec.Tree.Type)
	require.Len(t, spec.Tree.Children, 1)
	assert.Equal(t, "action", spec.Tree.Children[0].Type)
}

func TestNodeSpecWithParameters(t *testing.T) {
	node := NodeSpec{
		Type: "action",
		Name: "move_to",
		Node: "MoveToTarget",
		Parameters: map[string]interface{}{
			"speed": 5.0,
			"range": 10.0,
		},
	}

	assert.Equal(t, "action", node.Type)
	assert.Equal(t, "MoveToTarget", node.Node)
	assert.Equal(t, 5.0, node.Parameters["speed"])
}

func TestCustomNodeDef(t *testing.T) {
	custom := CustomNodeDef{
		Name:        "DetectEnemy",
		Type:        "condition",
		Description: "Detect nearby enemies",
		Category:    "perception",
		Parameters: []ParameterDef{
			{Name: "radius", Type: "float", Default: 15.0},
		},
		BlackboardReads:  []string{"position"},
		BlackboardWrites: []string{"target"},
	}

	assert.Equal(t, "DetectEnemy", custom.Name)
	assert.Equal(t, "condition", custom.Type)
	require.Len(t, custom.Parameters, 1)
	assert.Equal(t, "radius", custom.Parameters[0].Name)
}

func TestDecoratorTypes(t *testing.T) {
	decorators := BuiltinDecorators()
	require.NotEmpty(t, decorators)

	expectedNames := []string{"repeat", "negate", "always_succeed", "always_fail", "until_fail", "until_succeed", "timeout", "cooldown", "retry"}
	for _, name := range expectedNames {
		assert.Contains(t, decorators, name, "should contain decorator %s", name)
	}
}

func TestParallelPolicies(t *testing.T) {
	node := NodeSpec{
		Type:          "parallel",
		Name:          "watch_and_move",
		SuccessPolicy: "require_all",
		FailurePolicy: "require_one",
	}

	assert.Equal(t, "require_all", node.SuccessPolicy)
	assert.Equal(t, "require_one", node.FailurePolicy)
}

func TestSubtreeRef(t *testing.T) {
	ref := SubtreeRef{
		Name: "flee_subtree",
		File: "./subtrees/flee.beetree.yaml",
	}

	assert.Equal(t, "flee_subtree", ref.Name)
	assert.Equal(t, "./subtrees/flee.beetree.yaml", ref.File)
}

func TestExtensionNodeTypes(t *testing.T) {
	extensions := ExtensionNodeTypes()
	require.NotEmpty(t, extensions)
	assert.Contains(t, extensions, "utility_selector")
	assert.Contains(t, extensions, "active_selector")
	assert.Contains(t, extensions, "random_selector")
	assert.Contains(t, extensions, "random_sequence")
	assert.Contains(t, extensions, "subtree")
}

func TestIsValidNodeTypeIncludesExtensions(t *testing.T) {
	assert.True(t, IsValidNodeType("utility_selector"))
	assert.True(t, IsValidNodeType("active_selector"))
	assert.True(t, IsValidNodeType("subtree"))
}
