package validator

import (
	"testing"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_MinimalValid(t *testing.T) {
	spec := &model.TreeSpec{
		Version: "1.0",
		Metadata: model.Metadata{
			Name: "test",
		},
		Tree: model.NodeSpec{
			Type: "action",
			Name: "do_it",
			Node: "DoIt",
		},
	}

	errs := Validate(spec)
	assert.Empty(t, errs)
}

func TestValidate_MissingVersion(t *testing.T) {
	spec := &model.TreeSpec{
		Metadata: model.Metadata{Name: "test"},
		Tree:     model.NodeSpec{Type: "action", Name: "x", Node: "X"},
	}

	errs := Validate(spec)
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "version")
}

func TestValidate_MissingMetadataName(t *testing.T) {
	spec := &model.TreeSpec{
		Version: "1.0",
		Tree:    model.NodeSpec{Type: "action", Name: "x", Node: "X"},
	}

	errs := Validate(spec)
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "metadata.name")
}

func TestValidate_MissingTreeType(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree:     model.NodeSpec{Name: "root"},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "type")
}

func TestValidate_InvalidNodeType(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "bogus_type",
			Name: "root",
		},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "unknown node type")
}

func TestValidate_LeafWithChildren(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "action",
			Name: "root",
			Node: "DoIt",
			Children: []model.NodeSpec{
				{Type: "action", Name: "child", Node: "X"},
			},
		},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "leaf node")
}

func TestValidate_CompositeWithNoChildren(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "sequence",
			Name: "root",
		},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "must have at least one child")
}

func TestValidate_DecoratorMultipleChildren(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type:      "decorator",
			Name:      "root",
			Decorator: "negate",
			Children: []model.NodeSpec{
				{Type: "action", Name: "a", Node: "A"},
				{Type: "action", Name: "b", Node: "B"},
			},
		},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "exactly one child")
}

func TestValidate_DecoratorNoChildren(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type:      "decorator",
			Name:      "root",
			Decorator: "negate",
		},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "exactly one child")
}

func TestValidate_MissingNodeName(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "action",
			Node: "DoIt",
		},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "name is required")
}

func TestValidate_NestedErrors(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "selector",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "sequence", Name: "seq"}, // no children
				{Type: "bogus", Name: "bad"},    // invalid type
			},
		},
	}

	errs := Validate(spec)
	require.Len(t, errs, 2)
}

func TestValidate_ValidComplexTree(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "complex"},
		Blackboard: []model.BlackboardVar{
			{Name: "target", Type: "Entity"},
			{Name: "health", Type: "float"},
		},
		Tree: model.NodeSpec{
			Type: "selector",
			Name: "root",
			Children: []model.NodeSpec{
				{
					Type: "sequence",
					Name: "flee",
					Children: []model.NodeSpec{
						{Type: "condition", Name: "low_health", Node: "IsHealthLow"},
						{Type: "action", Name: "flee_action", Node: "Flee"},
					},
				},
				{
					Type: "decorator",
					Name: "repeat_patrol",
					Decorator: "repeat",
					Children: []model.NodeSpec{
						{Type: "action", Name: "patrol", Node: "Patrol"},
					},
				},
				{
					Type:          "parallel",
					Name:          "watch",
					SuccessPolicy: "require_all",
					FailurePolicy: "require_one",
					Children: []model.NodeSpec{
						{Type: "condition", Name: "safe", Node: "IsSafe"},
						{Type: "action", Name: "idle", Node: "Idle"},
					},
				},
			},
		},
	}

	errs := Validate(spec)
	assert.Empty(t, errs, "complex valid tree should have no errors: %v", errs)
}

func TestValidate_DuplicateBlackboardNames(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Blackboard: []model.BlackboardVar{
			{Name: "target", Type: "Entity"},
			{Name: "target", Type: "float"},
		},
		Tree: model.NodeSpec{Type: "action", Name: "x", Node: "X"},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "duplicate blackboard")
}

func TestValidate_CustomNodeInvalidType(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		CustomNodes: []model.CustomNodeDef{
			{Name: "MyNode", Type: "sequence"},
		},
		Tree: model.NodeSpec{Type: "action", Name: "x", Node: "X"},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "custom node")
}

func TestValidate_SubtreeDuplicateName(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree:     model.NodeSpec{Type: "action", Name: "do_it", Node: "DoIt"},
		Subtrees: []model.SubtreeRef{
			{Name: "patrol", File: "subtrees/patrol.beetree.yaml"},
			{Name: "patrol", File: "subtrees/patrol2.beetree.yaml"},
		},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "duplicate name")
}

func TestValidate_SubtreeMissingFile(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree:     model.NodeSpec{Type: "action", Name: "do_it", Node: "DoIt"},
		Subtrees: []model.SubtreeRef{
			{Name: "patrol", File: ""},
		},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "file is required")
}

func TestValidate_SubtreeRefNotFound(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "sequence",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "subtree", Name: "use_patrol", Ref: "nonexistent"},
			},
		},
	}

	errs := Validate(spec)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "subtree ref")
	assert.Contains(t, errs[0].Error(), "nonexistent")
}

func TestValidate_SubtreeRefValid(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "sequence",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "subtree", Name: "use_patrol", Ref: "patrol"},
			},
		},
		Subtrees: []model.SubtreeRef{
			{Name: "patrol", File: "subtrees/patrol.beetree.yaml"},
		},
	}

	errs := Validate(spec)
	assert.Empty(t, errs)
}
