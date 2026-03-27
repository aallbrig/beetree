package renderer

import (
	"testing"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderSpecASCII_SimpleAction(t *testing.T) {
	node := &model.NodeSpec{
		Type: "action",
		Name: "do_it",
	}

	output := RenderSpecASCII(node)
	assert.Contains(t, output, "do_it")
	assert.Contains(t, output, "ACT")
}

func TestRenderSpecASCII_SelectorWithChildren(t *testing.T) {
	node := &model.NodeSpec{
		Type: "selector",
		Name: "root",
		Children: []model.NodeSpec{
			{Type: "action", Name: "attack"},
			{Type: "action", Name: "patrol"},
		},
	}

	output := RenderSpecASCII(node)
	assert.Contains(t, output, "[SEL] root")
	assert.Contains(t, output, "[ACT] attack")
	assert.Contains(t, output, "[ACT] patrol")
}

func TestRenderSpecASCII_NestedTree(t *testing.T) {
	node := &model.NodeSpec{
		Type: "selector",
		Name: "root",
		Children: []model.NodeSpec{
			{
				Type: "sequence",
				Name: "engage",
				Children: []model.NodeSpec{
					{Type: "condition", Name: "has_target"},
					{Type: "action", Name: "attack"},
				},
			},
			{Type: "action", Name: "patrol"},
		},
	}

	output := RenderSpecASCII(node)
	assert.Contains(t, output, "SEL")
	assert.Contains(t, output, "SEQ")
	assert.Contains(t, output, "CND")
	assert.Contains(t, output, "ACT")
}

func TestRenderSpecASCII_Decorator(t *testing.T) {
	node := &model.NodeSpec{
		Type:      "decorator",
		Name:      "repeat_patrol",
		Decorator: "repeat",
		Children: []model.NodeSpec{
			{Type: "action", Name: "patrol"},
		},
	}

	output := RenderSpecASCII(node)
	assert.Contains(t, output, "DEC:repeat")
	assert.Contains(t, output, "repeat_patrol")
}

func TestRenderSpecASCII_Parallel(t *testing.T) {
	node := &model.NodeSpec{
		Type:          "parallel",
		Name:          "watch",
		SuccessPolicy: "require_all",
		Children: []model.NodeSpec{
			{Type: "condition", Name: "is_safe"},
			{Type: "action", Name: "idle"},
		},
	}

	output := RenderSpecASCII(node)
	assert.Contains(t, output, "PAR")
}

func TestRenderMermaid(t *testing.T) {
	node := &model.NodeSpec{
		Type: "selector",
		Name: "root",
		Children: []model.NodeSpec{
			{Type: "action", Name: "attack"},
			{Type: "action", Name: "patrol"},
		},
	}

	output := RenderMermaid(node)
	assert.Contains(t, output, "graph TD")
	assert.Contains(t, output, "root")
	assert.Contains(t, output, "attack")
	assert.Contains(t, output, "patrol")
	assert.Contains(t, output, "-->")
}

func TestRenderMermaid_NestedTree(t *testing.T) {
	node := &model.NodeSpec{
		Type: "selector",
		Name: "root",
		Children: []model.NodeSpec{
			{
				Type: "sequence",
				Name: "engage",
				Children: []model.NodeSpec{
					{Type: "condition", Name: "has_target"},
					{Type: "action", Name: "shoot"},
				},
			},
		},
	}

	output := RenderMermaid(node)
	require.Contains(t, output, "graph TD")
	assert.Contains(t, output, "engage")
	assert.Contains(t, output, "has_target")
	assert.Contains(t, output, "shoot")
}

func TestRenderDOT_SimpleTree(t *testing.T) {
	node := &model.NodeSpec{
		Type: "selector",
		Name: "root",
		Children: []model.NodeSpec{
			{Type: "action", Name: "attack"},
			{Type: "action", Name: "patrol"},
		},
	}

	output := RenderDOT(node)
	assert.Contains(t, output, "digraph")
	assert.Contains(t, output, "root")
	assert.Contains(t, output, "attack")
	assert.Contains(t, output, "patrol")
	assert.Contains(t, output, "->")
}

func TestRenderDOT_NestedTree(t *testing.T) {
	node := &model.NodeSpec{
		Type: "selector",
		Name: "root",
		Children: []model.NodeSpec{
			{
				Type: "sequence",
				Name: "engage",
				Children: []model.NodeSpec{
					{Type: "condition", Name: "has_target"},
					{Type: "action", Name: "shoot"},
				},
			},
		},
	}

	output := RenderDOT(node)
	assert.Contains(t, output, "digraph")
	assert.Contains(t, output, "engage")
	assert.Contains(t, output, "has_target")
	assert.Contains(t, output, "->")
}

func TestRenderDOT_Decorator(t *testing.T) {
	node := &model.NodeSpec{
		Type:      "decorator",
		Name:      "repeat_patrol",
		Decorator: "repeat",
		Children: []model.NodeSpec{
			{Type: "action", Name: "patrol"},
		},
	}

	output := RenderDOT(node)
	assert.Contains(t, output, "DEC:repeat")
}
