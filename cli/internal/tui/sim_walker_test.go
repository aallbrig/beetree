package tui

import (
	"testing"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func simSpec() *model.TreeSpec {
	return &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "sim-test"},
		Tree: model.NodeSpec{
			Type: "selector",
			Name: "root",
			Children: []model.NodeSpec{
				{
					Type: "sequence",
					Name: "combat",
					Children: []model.NodeSpec{
						{Type: "condition", Name: "has_target", Node: "HasTarget"},
						{Type: "action", Name: "attack", Node: "Attack"},
					},
				},
				{Type: "action", Name: "patrol", Node: "Patrol"},
			},
		},
	}
}

// --- Lifecycle ---

func TestSimWalker_NewStartsReady(t *testing.T) {
	sw := NewSimWalker(&simSpec().Tree)
	assert.Equal(t, SimReady, sw.State)
	assert.Nil(t, sw.CurrentNode)
	assert.Empty(t, sw.Trace)
}

func TestSimWalker_StepDescendsToFirstLeaf(t *testing.T) {
	sw := NewSimWalker(&simSpec().Tree)
	sw.Step()

	assert.Equal(t, SimWaitingForInput, sw.State)
	require.NotNil(t, sw.CurrentNode)
	assert.Equal(t, "has_target", sw.CurrentNode.Name)
	assert.Equal(t, "condition", sw.CurrentNode.Type)
}

// --- Sequence semantics: all children must succeed ---

func TestSimWalker_SequenceAllSucceed(t *testing.T) {
	// selector > sequence(has_target, attack) > patrol
	// If has_target=SUCCESS, attack=SUCCESS → sequence=SUCCESS → selector=SUCCESS
	sw := NewSimWalker(&simSpec().Tree)

	sw.Step()
	assert.Equal(t, "has_target", sw.CurrentNode.Name)
	sw.Resolve(model.StatusSuccess)

	// Should advance to next child in sequence
	assert.Equal(t, SimWaitingForInput, sw.State)
	assert.Equal(t, "attack", sw.CurrentNode.Name)
	sw.Resolve(model.StatusSuccess)

	// Sequence succeeds → selector succeeds → complete
	assert.Equal(t, SimComplete, sw.State)
	assert.Equal(t, model.StatusSuccess, sw.Result)
}

func TestSimWalker_SequenceFirstChildFails(t *testing.T) {
	// has_target=FAILURE → sequence=FAILURE → selector tries patrol
	sw := NewSimWalker(&simSpec().Tree)

	sw.Step()
	assert.Equal(t, "has_target", sw.CurrentNode.Name)
	sw.Resolve(model.StatusFailure)

	// Sequence fails, selector moves to patrol
	assert.Equal(t, SimWaitingForInput, sw.State)
	assert.Equal(t, "patrol", sw.CurrentNode.Name)
}

func TestSimWalker_SequenceSecondChildFails(t *testing.T) {
	// has_target=SUCCESS, attack=FAILURE → sequence=FAILURE → selector tries patrol
	sw := NewSimWalker(&simSpec().Tree)

	sw.Step()
	sw.Resolve(model.StatusSuccess) // has_target

	assert.Equal(t, "attack", sw.CurrentNode.Name)
	sw.Resolve(model.StatusFailure) // attack fails

	// Selector moves to patrol
	assert.Equal(t, SimWaitingForInput, sw.State)
	assert.Equal(t, "patrol", sw.CurrentNode.Name)
}

// --- Selector semantics: first success wins ---

func TestSimWalker_SelectorFirstChildSucceeds(t *testing.T) {
	sw := NewSimWalker(&simSpec().Tree)

	sw.Step()
	sw.Resolve(model.StatusSuccess) // has_target
	sw.Resolve(model.StatusSuccess) // attack

	// Selector got SUCCESS from sequence → done
	assert.Equal(t, SimComplete, sw.State)
	assert.Equal(t, model.StatusSuccess, sw.Result)
}

func TestSimWalker_SelectorAllChildrenFail(t *testing.T) {
	sw := NewSimWalker(&simSpec().Tree)

	sw.Step()
	sw.Resolve(model.StatusFailure) // has_target → sequence fails

	assert.Equal(t, "patrol", sw.CurrentNode.Name)
	sw.Resolve(model.StatusFailure) // patrol fails

	// All selector children failed → selector fails
	assert.Equal(t, SimComplete, sw.State)
	assert.Equal(t, model.StatusFailure, sw.Result)
}

func TestSimWalker_SelectorFallbackSucceeds(t *testing.T) {
	sw := NewSimWalker(&simSpec().Tree)

	sw.Step()
	sw.Resolve(model.StatusFailure) // has_target → sequence fails

	assert.Equal(t, "patrol", sw.CurrentNode.Name)
	sw.Resolve(model.StatusSuccess) // patrol succeeds

	assert.Equal(t, SimComplete, sw.State)
	assert.Equal(t, model.StatusSuccess, sw.Result)
}

// --- RUNNING status ---

func TestSimWalker_RunningBubblesUp(t *testing.T) {
	sw := NewSimWalker(&simSpec().Tree)

	sw.Step()
	sw.Resolve(model.StatusRunning) // has_target returns RUNNING

	// RUNNING in sequence → sequence returns RUNNING → selector returns RUNNING
	assert.Equal(t, SimComplete, sw.State)
	assert.Equal(t, model.StatusRunning, sw.Result)
}

// --- Execution trace ---

func TestSimWalker_TraceRecordsAllSteps(t *testing.T) {
	sw := NewSimWalker(&simSpec().Tree)

	sw.Step()
	sw.Resolve(model.StatusSuccess) // has_target
	sw.Resolve(model.StatusSuccess) // attack

	require.NotEmpty(t, sw.Trace)

	// Should have entries for: enter root, enter combat, resolve has_target,
	// resolve attack, resolve combat, resolve root
	traceNames := make([]string, len(sw.Trace))
	for i, s := range sw.Trace {
		traceNames[i] = s.NodeName
	}
	assert.Contains(t, traceNames, "root")
	assert.Contains(t, traceNames, "combat")
	assert.Contains(t, traceNames, "has_target")
	assert.Contains(t, traceNames, "attack")
}

func TestSimWalker_TraceShowsEnterAndResolve(t *testing.T) {
	sw := NewSimWalker(&simSpec().Tree)
	sw.Step()
	sw.Resolve(model.StatusSuccess) // has_target
	sw.Resolve(model.StatusSuccess) // attack

	// Find leaf resolve events
	var leafResolves []SimStep
	for _, s := range sw.Trace {
		if s.Event == "resolve" && (s.NodeType == "action" || s.NodeType == "condition") {
			leafResolves = append(leafResolves, s)
		}
	}
	require.Len(t, leafResolves, 2)
	assert.Equal(t, "has_target", leafResolves[0].NodeName)
	assert.Equal(t, model.StatusSuccess, leafResolves[0].Status)
	assert.Equal(t, "attack", leafResolves[1].NodeName)
	assert.Equal(t, model.StatusSuccess, leafResolves[1].Status)
}

// --- Decorator handling ---

func TestSimWalker_DecoratorNegate(t *testing.T) {
	tree := &model.NodeSpec{
		Type: "sequence",
		Name: "root",
		Children: []model.NodeSpec{
			{
				Type:      "condition",
				Name:      "not_dead",
				Node:      "IsDead",
				Decorator: "negate",
			},
			{Type: "action", Name: "act", Node: "DoThing"},
		},
	}

	sw := NewSimWalker(tree)
	sw.Step()

	assert.Equal(t, "not_dead", sw.CurrentNode.Name)
	sw.Resolve(model.StatusSuccess) // user says SUCCESS, negate flips to FAILURE

	// Negated: SUCCESS→FAILURE, so sequence fails
	// Root sequence fails → complete
	assert.Equal(t, SimComplete, sw.State)
	assert.Equal(t, model.StatusFailure, sw.Result)
}

func TestSimWalker_DecoratorAlwaysSucceed(t *testing.T) {
	tree := &model.NodeSpec{
		Type: "sequence",
		Name: "root",
		Children: []model.NodeSpec{
			{
				Type:      "action",
				Name:      "risky",
				Node:      "RiskyAction",
				Decorator: "always_succeed",
			},
			{Type: "action", Name: "next", Node: "NextAction"},
		},
	}

	sw := NewSimWalker(tree)
	sw.Step()

	assert.Equal(t, "risky", sw.CurrentNode.Name)
	sw.Resolve(model.StatusFailure) // user says FAILURE, always_succeed overrides

	// always_succeed: FAILURE→SUCCESS, so sequence continues
	assert.Equal(t, SimWaitingForInput, sw.State)
	assert.Equal(t, "next", sw.CurrentNode.Name)
}

func TestSimWalker_DecoratorAlwaysFail(t *testing.T) {
	tree := &model.NodeSpec{
		Type: "selector",
		Name: "root",
		Children: []model.NodeSpec{
			{
				Type:      "action",
				Name:      "check",
				Decorator: "always_fail",
			},
			{Type: "action", Name: "fallback"},
		},
	}

	sw := NewSimWalker(tree)
	sw.Step()
	sw.Resolve(model.StatusSuccess) // always_fail overrides to FAILURE

	// Selector moves to fallback
	assert.Equal(t, SimWaitingForInput, sw.State)
	assert.Equal(t, "fallback", sw.CurrentNode.Name)
}

// --- Single leaf tree ---

func TestSimWalker_SingleLeafTree(t *testing.T) {
	tree := &model.NodeSpec{
		Type: "action",
		Name: "only_node",
		Node: "DoSomething",
	}

	sw := NewSimWalker(tree)
	sw.Step()

	assert.Equal(t, SimWaitingForInput, sw.State)
	assert.Equal(t, "only_node", sw.CurrentNode.Name)

	sw.Resolve(model.StatusSuccess)
	assert.Equal(t, SimComplete, sw.State)
	assert.Equal(t, model.StatusSuccess, sw.Result)
}

// --- Parallel node ---

func TestSimWalker_ParallelAsksAllChildren(t *testing.T) {
	tree := &model.NodeSpec{
		Type: "parallel",
		Name: "root",
		Children: []model.NodeSpec{
			{Type: "action", Name: "move", Node: "Move"},
			{Type: "action", Name: "shoot", Node: "Shoot"},
		},
	}

	sw := NewSimWalker(tree)
	sw.Step()

	// Parallel evaluates all children, so first child
	assert.Equal(t, "move", sw.CurrentNode.Name)
	sw.Resolve(model.StatusSuccess)

	// Then second child
	assert.Equal(t, SimWaitingForInput, sw.State)
	assert.Equal(t, "shoot", sw.CurrentNode.Name)
	sw.Resolve(model.StatusSuccess)

	// Both succeed → parallel succeeds (default: all must succeed)
	assert.Equal(t, SimComplete, sw.State)
	assert.Equal(t, model.StatusSuccess, sw.Result)
}

func TestSimWalker_ParallelOneFails(t *testing.T) {
	tree := &model.NodeSpec{
		Type: "parallel",
		Name: "root",
		Children: []model.NodeSpec{
			{Type: "action", Name: "move"},
			{Type: "action", Name: "shoot"},
		},
	}

	sw := NewSimWalker(tree)
	sw.Step()
	sw.Resolve(model.StatusSuccess) // move

	sw.Resolve(model.StatusFailure) // shoot fails

	// Default parallel: any failure = failure
	assert.Equal(t, SimComplete, sw.State)
	assert.Equal(t, model.StatusFailure, sw.Result)
}

// --- Deep nesting ---

func TestSimWalker_DeepNesting(t *testing.T) {
	tree := &model.NodeSpec{
		Type: "selector",
		Name: "root",
		Children: []model.NodeSpec{
			{
				Type: "sequence",
				Name: "branch1",
				Children: []model.NodeSpec{
					{
						Type: "selector",
						Name: "sub_select",
						Children: []model.NodeSpec{
							{Type: "condition", Name: "deep_check"},
							{Type: "action", Name: "deep_act"},
						},
					},
					{Type: "action", Name: "after_sub"},
				},
			},
			{Type: "action", Name: "fallback"},
		},
	}

	sw := NewSimWalker(tree)
	sw.Step()

	// Should descend: root → branch1 → sub_select → deep_check
	assert.Equal(t, "deep_check", sw.CurrentNode.Name)

	sw.Resolve(model.StatusFailure) // deep_check fails

	// sub_select tries deep_act
	assert.Equal(t, "deep_act", sw.CurrentNode.Name)
	sw.Resolve(model.StatusSuccess) // deep_act succeeds

	// sub_select = SUCCESS → sequence continues to after_sub
	assert.Equal(t, "after_sub", sw.CurrentNode.Name)
	sw.Resolve(model.StatusSuccess)

	// sequence = SUCCESS → selector = SUCCESS → done
	assert.Equal(t, SimComplete, sw.State)
	assert.Equal(t, model.StatusSuccess, sw.Result)
}

// --- Resolve without Step is no-op ---

func TestSimWalker_ResolveWhenNotWaiting(t *testing.T) {
	sw := NewSimWalker(&simSpec().Tree)
	// Resolve before Step — should not panic or change state
	sw.Resolve(model.StatusSuccess)
	assert.Equal(t, SimReady, sw.State)
}

// --- Reset ---

func TestSimWalker_Reset(t *testing.T) {
	sw := NewSimWalker(&simSpec().Tree)
	sw.Step()
	sw.Resolve(model.StatusSuccess)
	sw.Resolve(model.StatusSuccess)
	require.Equal(t, SimComplete, sw.State)

	sw.Reset()
	assert.Equal(t, SimReady, sw.State)
	assert.Nil(t, sw.CurrentNode)
	assert.Empty(t, sw.Trace)
	assert.Empty(t, sw.stack)
}
