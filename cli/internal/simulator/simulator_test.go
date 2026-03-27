package simulator

import (
	"strings"
	"testing"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimulate_SingleAction(t *testing.T) {
	spec := &model.TreeSpec{
		Tree: model.NodeSpec{
			Type: "action",
			Name: "patrol",
			Node: "Patrol",
		},
	}

	result := Simulate(spec, nil)
	require.NotNil(t, result)
	assert.Equal(t, StatusSuccess, result.Status)
	assert.Equal(t, "patrol", result.NodeName)
	assert.Len(t, result.Steps, 1)
}

func TestSimulate_Sequence_AllSucceed(t *testing.T) {
	spec := &model.TreeSpec{
		Tree: model.NodeSpec{
			Type: "sequence",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "action", Name: "step1", Node: "Step1"},
				{Type: "action", Name: "step2", Node: "Step2"},
			},
		},
	}

	result := Simulate(spec, nil)
	assert.Equal(t, StatusSuccess, result.Status)
	assert.Len(t, result.Steps, 4) // enter sequence, action1, action2, exit sequence
}

func TestSimulate_Sequence_ChildFails(t *testing.T) {
	spec := &model.TreeSpec{
		Tree: model.NodeSpec{
			Type: "sequence",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "action", Name: "step1", Node: "Step1"},
				{Type: "condition", Name: "check", Node: "Check"},
				{Type: "action", Name: "step2", Node: "Step2"},
			},
		},
	}

	// Make the condition fail
	overrides := map[string]Status{"check": StatusFailure}
	result := Simulate(spec, overrides)
	assert.Equal(t, StatusFailure, result.Status)
	// step2 should not execute
	for _, s := range result.Steps {
		assert.NotEqual(t, "step2", s.NodeName)
	}
}

func TestSimulate_Selector_FirstFails(t *testing.T) {
	spec := &model.TreeSpec{
		Tree: model.NodeSpec{
			Type: "selector",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "action", Name: "option1", Node: "Option1"},
				{Type: "action", Name: "option2", Node: "Option2"},
			},
		},
	}

	overrides := map[string]Status{"option1": StatusFailure}
	result := Simulate(spec, overrides)
	assert.Equal(t, StatusSuccess, result.Status)
	// option2 should execute because option1 failed
	found := false
	for _, s := range result.Steps {
		if s.NodeName == "option2" {
			found = true
		}
	}
	assert.True(t, found, "option2 should have been reached")
}

func TestSimulate_Selector_AllFail(t *testing.T) {
	spec := &model.TreeSpec{
		Tree: model.NodeSpec{
			Type: "selector",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "action", Name: "a", Node: "A"},
				{Type: "action", Name: "b", Node: "B"},
			},
		},
	}

	overrides := map[string]Status{"a": StatusFailure, "b": StatusFailure}
	result := Simulate(spec, overrides)
	assert.Equal(t, StatusFailure, result.Status)
}

func TestSimulate_Decorator_Negate(t *testing.T) {
	spec := &model.TreeSpec{
		Tree: model.NodeSpec{
			Type:      "decorator",
			Name:      "not_check",
			Decorator: "negate",
			Children: []model.NodeSpec{
				{Type: "condition", Name: "check", Node: "Check"},
			},
		},
	}

	// Check succeeds, negate inverts to failure
	result := Simulate(spec, nil)
	assert.Equal(t, StatusFailure, result.Status)
}

func TestSimulate_FormatTrace(t *testing.T) {
	spec := &model.TreeSpec{
		Tree: model.NodeSpec{
			Type: "sequence",
			Name: "root",
			Children: []model.NodeSpec{
				{Type: "action", Name: "walk", Node: "Walk"},
			},
		},
	}

	result := Simulate(spec, nil)
	trace := FormatTrace(result)
	assert.Contains(t, trace, "root")
	assert.Contains(t, trace, "walk")
	assert.Contains(t, trace, "SUCCESS")
}

func TestSimulate_Parallel(t *testing.T) {
	spec := &model.TreeSpec{
		Tree: model.NodeSpec{
			Type: "parallel",
			Name: "both",
			Children: []model.NodeSpec{
				{Type: "action", Name: "move", Node: "Move"},
				{Type: "action", Name: "shoot", Node: "Shoot"},
			},
		},
	}

	result := Simulate(spec, nil)
	assert.Equal(t, StatusSuccess, result.Status)
}

func TestSimulate_NestedTree(t *testing.T) {
	spec := &model.TreeSpec{
		Tree: model.NodeSpec{
			Type: "selector",
			Name: "ai_root",
			Children: []model.NodeSpec{
				{
					Type: "sequence",
					Name: "flee",
					Children: []model.NodeSpec{
						{Type: "condition", Name: "low_health", Node: "LowHealth"},
						{Type: "action", Name: "run_away", Node: "RunAway"},
					},
				},
				{
					Type: "action",
					Name: "patrol",
					Node: "Patrol",
				},
			},
		},
	}

	// low_health succeeds -> flee sequence runs -> run_away succeeds -> selector succeeds (first child)
	result := Simulate(spec, nil)
	assert.Equal(t, StatusSuccess, result.Status)
	trace := FormatTrace(result)
	assert.True(t, strings.Contains(trace, "low_health"), "should show condition check")
	assert.True(t, strings.Contains(trace, "run_away"), "should execute run_away")
	// patrol should NOT execute since flee succeeded
	for _, s := range result.Steps {
		if s.NodeName == "patrol" && s.Event == "execute" {
			t.Error("patrol should not execute when flee succeeds")
		}
	}
}
