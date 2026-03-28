package renderer

import (
	"testing"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSigil(t *testing.T) {
	empty := model.NotationConfig{}

	tests := []struct {
		name     string
		node     model.NodeSpec
		notation model.NotationConfig
		expected string
	}{
		{"action", model.NodeSpec{Type: "action"}, empty, "!"},
		{"condition", model.NodeSpec{Type: "condition"}, empty, "¿"},
		{"sequence", model.NodeSpec{Type: "sequence"}, empty, "→"},
		{"selector", model.NodeSpec{Type: "selector"}, empty, "?"},
		{"parallel", model.NodeSpec{Type: "parallel"}, empty, "⇒"},
		{"decorator generic", model.NodeSpec{Type: "decorator", Decorator: "custom_dec"}, empty, "◇custom_dec"},
		{"decorator repeat", model.NodeSpec{Type: "decorator", Decorator: "repeat"}, empty, "◇∞"},
		{"decorator negate", model.NodeSpec{Type: "decorator", Decorator: "negate"}, empty, "◇¬"},
		{"decorator always_succeed", model.NodeSpec{Type: "decorator", Decorator: "always_succeed"}, empty, "◇✓"},
		{"decorator always_fail", model.NodeSpec{Type: "decorator", Decorator: "always_fail"}, empty, "◇✗"},
		{"decorator retry", model.NodeSpec{Type: "decorator", Decorator: "retry"}, empty, "◇↻"},
		{"subtree", model.NodeSpec{Type: "subtree"}, empty, "↗"},
		{"utility_selector", model.NodeSpec{Type: "utility_selector"}, empty, "⚖"},
		{"active_selector", model.NodeSpec{Type: "active_selector"}, empty, "⚡"},
		{"random_selector", model.NodeSpec{Type: "random_selector"}, empty, "?~"},
		{"random_sequence", model.NodeSpec{Type: "random_sequence"}, empty, "→~"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ResolveSigil(&tt.node, tt.notation))
		})
	}
}

func TestResolveSigil_CustomOverrides(t *testing.T) {
	t.Run("type sigil override", func(t *testing.T) {
		notation := model.NotationConfig{
			TypeSigils: map[string]string{
				"sequence": "»",
			},
		}
		node := &model.NodeSpec{Type: "sequence", Name: "patrol"}
		assert.Equal(t, "»", ResolveSigil(node, notation))
	})

	t.Run("node class sigil override", func(t *testing.T) {
		notation := model.NotationConfig{
			NodeSigils: map[string]string{
				"PlayAnimation": "▶",
			},
		}
		node := &model.NodeSpec{Type: "action", Name: "play_anim", Node: "PlayAnimation"}
		assert.Equal(t, "▶", ResolveSigil(node, notation))
	})

	t.Run("node class takes priority over type override", func(t *testing.T) {
		notation := model.NotationConfig{
			TypeSigils: map[string]string{
				"action": "A",
			},
			NodeSigils: map[string]string{
				"PlayAnimation": "▶",
			},
		}
		node := &model.NodeSpec{Type: "action", Name: "play_anim", Node: "PlayAnimation"}
		assert.Equal(t, "▶", ResolveSigil(node, notation))
	})

	t.Run("type override does not affect other types", func(t *testing.T) {
		notation := model.NotationConfig{
			TypeSigils: map[string]string{
				"sequence": "»",
			},
		}
		node := &model.NodeSpec{Type: "selector", Name: "root"}
		assert.Equal(t, "?", ResolveSigil(node, notation))
	})
}

func TestRenderSigil(t *testing.T) {
	tree := &model.NodeSpec{
		Type: "selector",
		Name: "root",
		Children: []model.NodeSpec{
			{
				Type: "sequence",
				Name: "patrol",
				Children: []model.NodeSpec{
					{Type: "condition", Name: "has_waypoint"},
					{Type: "action", Name: "move_to"},
				},
			},
			{Type: "action", Name: "idle"},
		},
	}

	output := RenderSigil(tree, model.NotationConfig{})

	assert.Contains(t, output, "? root")
	assert.Contains(t, output, "→ patrol")
	assert.Contains(t, output, "¿ has_waypoint")
	assert.Contains(t, output, "! move_to")
	assert.Contains(t, output, "! idle")
	assert.Contains(t, output, "├── ")
	assert.Contains(t, output, "└── ")
}

func TestRenderCompact(t *testing.T) {
	tree := &model.NodeSpec{
		Type: "selector",
		Name: "root",
		Children: []model.NodeSpec{
			{
				Type: "sequence",
				Name: "patrol",
				Children: []model.NodeSpec{
					{Type: "condition", Name: "has_waypoint"},
					{Type: "action", Name: "move_to"},
				},
			},
		},
	}

	output := RenderCompact(tree, model.NotationConfig{})

	expected := "? root\n  → patrol\n    ¿ has_waypoint\n    ! move_to\n"
	assert.Equal(t, expected, output)
}

func TestRenderOneline(t *testing.T) {
	tree := &model.NodeSpec{
		Type: "selector",
		Name: "root",
		Children: []model.NodeSpec{
			{
				Type: "sequence",
				Name: "patrol",
				Children: []model.NodeSpec{
					{Type: "condition", Name: "has_wp"},
					{Type: "action", Name: "move"},
				},
			},
			{Type: "action", Name: "idle"},
		},
	}

	output := RenderOneline(tree, model.NotationConfig{})
	assert.Equal(t, "?root(→patrol(¿has_wp !move) !idle)", output)
}

func TestRenderSigil_WithDecorator(t *testing.T) {
	tree := &model.NodeSpec{
		Type:      "decorator",
		Name:      "loop",
		Decorator: "repeat",
		Children: []model.NodeSpec{
			{
				Type: "sequence",
				Name: "cycle",
				Children: []model.NodeSpec{
					{Type: "action", Name: "walk"},
					{Type: "action", Name: "wait"},
				},
			},
		},
	}

	output := RenderSigil(tree, model.NotationConfig{})
	require.Contains(t, output, "◇∞ loop")
	assert.Contains(t, output, "→ cycle")
}

func TestRenderSigil_WithCustomNotation(t *testing.T) {
	notation := model.NotationConfig{
		NodeSigils: map[string]string{
			"PlayAnim": "▶",
		},
	}

	tree := &model.NodeSpec{
		Type: "sequence",
		Name: "anim_seq",
		Children: []model.NodeSpec{
			{Type: "action", Name: "play_walk", Node: "PlayAnim"},
			{Type: "action", Name: "play_idle", Node: "PlayAnim"},
			{Type: "action", Name: "shoot"},
		},
	}

	output := RenderSigil(tree, notation)
	assert.Contains(t, output, "▶ play_walk")
	assert.Contains(t, output, "▶ play_idle")
	assert.Contains(t, output, "! shoot")
}
