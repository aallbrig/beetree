package tui

import (
	"testing"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNodeTypeTag(t *testing.T) {
	tests := []struct {
		nodeType string
		expected string
	}{
		{"selector", "[SEL]"},
		{"sequence", "[SEQ]"},
		{"parallel", "[PAR]"},
		{"action", "[ACT]"},
		{"condition", "[CND]"},
		{"decorator", "[DEC]"},
		{"utility_selector", "[UTL]"},
		{"active_selector", "[ASL]"},
		{"random_selector", "[RND]"},
		{"random_sequence", "[RSQ]"},
		{"subtree", "[SUB]"},
		{"unknown_type", "[???]"},
		{"", "[???]"},
	}

	for _, tt := range tests {
		t.Run(tt.nodeType, func(t *testing.T) {
			assert.Equal(t, tt.expected, NodeTypeTag(tt.nodeType))
		})
	}
}

func TestNodeLabel(t *testing.T) {
	tests := []struct {
		name     string
		node     model.NodeSpec
		expected string
	}{
		{
			name:     "selector node",
			node:     model.NodeSpec{Type: "selector", Name: "root"},
			expected: "? root",
		},
		{
			name:     "action with class ref",
			node:     model.NodeSpec{Type: "action", Name: "attack", Node: "PlayerAttack"},
			expected: "! attack (PlayerAttack)",
		},
		{
			name:     "condition with class ref",
			node:     model.NodeSpec{Type: "condition", Name: "has_target", Node: "HasTarget"},
			expected: "¿ has_target (HasTarget)",
		},
		{
			name:     "decorated node shows decorator sigil",
			node:     model.NodeSpec{Type: "decorator", Name: "retry_attack", Decorator: "retry"},
			expected: "◇↻ retry_attack",
		},
		{
			name:     "subtree ref",
			node:     model.NodeSpec{Type: "subtree", Name: "patrol_sub", Ref: "patrol-tree"},
			expected: "↗ patrol_sub →patrol-tree",
		},
		{
			name:     "action no extras",
			node:     model.NodeSpec{Type: "action", Name: "idle"},
			expected: "! idle",
		},
		{
			name:     "sequence node",
			node:     model.NodeSpec{Type: "sequence", Name: "patrol_cycle"},
			expected: "→ patrol_cycle",
		},
		{
			name:     "parallel node",
			node:     model.NodeSpec{Type: "parallel", Name: "engage"},
			expected: "⇒ engage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NodeLabel(&tt.node))
		})
	}
}

func TestNodeProperties(t *testing.T) {
	t.Run("basic node", func(t *testing.T) {
		node := &model.NodeSpec{
			Type: "selector",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "action", Name: "a1"},
				{Type: "action", Name: "a2"},
			},
		}
		props := NodeProperties(node)
		assert.Equal(t, "root", props.Name)
		assert.Equal(t, "selector", props.Type)
		assert.Equal(t, 2, props.ChildCount)
		assert.Empty(t, props.NodeClass)
		assert.Empty(t, props.Decorator)
	})

	t.Run("leaf with class and decorator", func(t *testing.T) {
		node := &model.NodeSpec{
			Type:      "action",
			Name:      "attack",
			Node:      "PlayerAttack",
			Decorator: "retry",
			Parameters: map[string]interface{}{
				"damage": 10,
				"range":  5.0,
			},
		}
		props := NodeProperties(node)
		assert.Equal(t, "attack", props.Name)
		assert.Equal(t, "action", props.Type)
		assert.Equal(t, 0, props.ChildCount)
		assert.Equal(t, "PlayerAttack", props.NodeClass)
		assert.Equal(t, "retry", props.Decorator)
		assert.Len(t, props.Parameters, 2)
	})

	t.Run("subtree reference", func(t *testing.T) {
		node := &model.NodeSpec{
			Type: "subtree",
			Name: "patrol_sub",
			Ref:  "patrol-tree",
			File: "trees/patrol.beetree.yaml",
		}
		props := NodeProperties(node)
		assert.Equal(t, "patrol-tree", props.SubtreeRef)
		assert.Equal(t, "trees/patrol.beetree.yaml", props.SubtreeFile)
	})

	t.Run("nil node returns empty", func(t *testing.T) {
		props := NodeProperties(nil)
		assert.Equal(t, "", props.Name)
		assert.Equal(t, "", props.Type)
	})
}

func TestAvailableNodeTypes(t *testing.T) {
	types := AvailableNodeTypes()

	// Should contain core types
	found := map[string]bool{}
	for _, nt := range types {
		found[nt.Name] = true
	}
	assert.True(t, found["selector"])
	assert.True(t, found["sequence"])
	assert.True(t, found["action"])
	assert.True(t, found["condition"])
	assert.True(t, found["parallel"])
	assert.True(t, found["decorator"])

	// Should contain extension types
	assert.True(t, found["utility_selector"])
	assert.True(t, found["subtree"])

	// Each entry should have a category
	for _, nt := range types {
		assert.NotEmpty(t, nt.Category, "type %s should have category", nt.Name)
	}
}

func TestFilterNodeTypes(t *testing.T) {
	all := AvailableNodeTypes()

	t.Run("empty filter returns all", func(t *testing.T) {
		filtered := FilterNodeTypes(all, "")
		assert.Equal(t, len(all), len(filtered))
	})

	t.Run("filter by prefix", func(t *testing.T) {
		filtered := FilterNodeTypes(all, "sel")
		assert.GreaterOrEqual(t, len(filtered), 1)
		for _, nt := range filtered {
			assert.Contains(t, nt.Name, "sel")
		}
	})

	t.Run("filter case insensitive", func(t *testing.T) {
		filtered := FilterNodeTypes(all, "SEQ")
		assert.GreaterOrEqual(t, len(filtered), 1)
	})

	t.Run("no matches", func(t *testing.T) {
		filtered := FilterNodeTypes(all, "zzzzz")
		assert.Empty(t, filtered)
	})
}
