package differ

import (
	"strings"
	"testing"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiff_Identical(t *testing.T) {
	spec := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree:     model.NodeSpec{Type: "action", Name: "patrol", Node: "Patrol"},
	}
	changes := Diff(spec, spec)
	assert.Empty(t, changes)
}

func TestDiff_MetadataChanged(t *testing.T) {
	a := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test", Description: "old desc"},
		Tree:     model.NodeSpec{Type: "action", Name: "patrol", Node: "Patrol"},
	}
	b := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test", Description: "new desc"},
		Tree:     model.NodeSpec{Type: "action", Name: "patrol", Node: "Patrol"},
	}
	changes := Diff(a, b)
	require.NotEmpty(t, changes)
	assert.Contains(t, changes[0].Path, "metadata.description")
	assert.Equal(t, ChangeModified, changes[0].Type)
}

func TestDiff_NodeAdded(t *testing.T) {
	a := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "sequence",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "action", Name: "step1", Node: "Step1"},
			},
		},
	}
	b := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "sequence",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "action", Name: "step1", Node: "Step1"},
				{Type: "action", Name: "step2", Node: "Step2"},
			},
		},
	}
	changes := Diff(a, b)
	require.NotEmpty(t, changes)
	found := false
	for _, c := range changes {
		if c.Type == ChangeAdded && strings.Contains(c.Path, "step2") {
			found = true
		}
	}
	assert.True(t, found, "should detect added node step2")
}

func TestDiff_NodeRemoved(t *testing.T) {
	a := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "sequence",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "action", Name: "step1", Node: "Step1"},
				{Type: "action", Name: "step2", Node: "Step2"},
			},
		},
	}
	b := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "sequence",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "action", Name: "step1", Node: "Step1"},
			},
		},
	}
	changes := Diff(a, b)
	require.NotEmpty(t, changes)
	found := false
	for _, c := range changes {
		if c.Type == ChangeRemoved && strings.Contains(c.Path, "step2") {
			found = true
		}
	}
	assert.True(t, found, "should detect removed node step2")
}

func TestDiff_BlackboardChanged(t *testing.T) {
	a := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Blackboard: []model.BlackboardVar{
			{Name: "health", Type: "float", Default: 100.0},
		},
		Tree: model.NodeSpec{Type: "action", Name: "patrol", Node: "Patrol"},
	}
	b := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Blackboard: []model.BlackboardVar{
			{Name: "health", Type: "float", Default: 50.0},
		},
		Tree: model.NodeSpec{Type: "action", Name: "patrol", Node: "Patrol"},
	}
	changes := Diff(a, b)
	require.NotEmpty(t, changes)
	assert.Contains(t, changes[0].Path, "blackboard")
	assert.Contains(t, changes[0].Path, "health")
}

func TestDiff_NodeTypeChanged(t *testing.T) {
	a := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree:     model.NodeSpec{Type: "action", Name: "patrol", Node: "Patrol"},
	}
	b := &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree:     model.NodeSpec{Type: "condition", Name: "patrol", Node: "Patrol"},
	}
	changes := Diff(a, b)
	require.NotEmpty(t, changes)
	assert.Equal(t, ChangeModified, changes[0].Type)
	assert.Contains(t, changes[0].Path, "tree.type")
}

func TestFormatDiff(t *testing.T) {
	changes := []Change{
		{Type: ChangeAdded, Path: "tree.children[1]", New: "step2 (action)"},
		{Type: ChangeRemoved, Path: "blackboard.target", Old: "Entity"},
		{Type: ChangeModified, Path: "metadata.description", Old: "old", New: "new"},
	}
	output := FormatDiff(changes)
	assert.Contains(t, output, "+")
	assert.Contains(t, output, "-")
	assert.Contains(t, output, "~")
}
