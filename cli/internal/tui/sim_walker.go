package tui

import "github.com/aallbrig/beetree-cli/internal/model"

// SimState represents the simulation lifecycle.
type SimState int

const (
	SimReady           SimState = iota // not started
	SimWaitingForInput                 // paused on a leaf, waiting for user
	SimComplete                        // tree fully evaluated
)

// SimStep records one event in the simulation trace.
type SimStep struct {
	NodeName string
	NodeType string
	Status   model.Status
	Depth    int
	Event    string // "enter", "resolve", "skip"
}

// frameEntry tracks evaluation state for a composite node on the stack.
type frameEntry struct {
	node       *model.NodeSpec
	childIndex int            // next child to evaluate
	results    []model.Status // results collected so far
	depth      int
}

// SimWalker walks a behavior tree interactively, pausing at each leaf
// for the user to choose the return status (Success/Failure/Running).
// Composite nodes resolve automatically from their children's results.
type SimWalker struct {
	root        *model.NodeSpec
	State       SimState
	CurrentNode *model.NodeSpec
	Trace       []SimStep
	Result      model.Status
	stack       []frameEntry
}

// NewSimWalker creates a walker for the given tree root.
func NewSimWalker(root *model.NodeSpec) *SimWalker {
	return &SimWalker{
		root:  root,
		State: SimReady,
	}
}

// Step begins or continues the simulation, descending until a leaf node
// needs user input. Call this once to start, then use Resolve() for each leaf.
func (sw *SimWalker) Step() {
	if sw.State == SimComplete {
		return
	}
	if sw.State == SimReady {
		sw.descend(sw.root, 0)
		return
	}
}

// Resolve provides the user's chosen status for the current leaf node,
// then continues execution until the next leaf or completion.
func (sw *SimWalker) Resolve(status model.Status) {
	if sw.State != SimWaitingForInput {
		return
	}

	node := sw.CurrentNode

	// Apply decorator transformation
	resolved := applyDecorator(node, status)

	sw.Trace = append(sw.Trace, SimStep{
		NodeName: node.Name,
		NodeType: node.Type,
		Status:   resolved,
		Depth:    len(sw.stack),
		Event:    "resolve",
	})

	sw.CurrentNode = nil
	sw.bubbleUp(resolved)
}

// Reset returns the walker to its initial state for re-running.
func (sw *SimWalker) Reset() {
	sw.State = SimReady
	sw.CurrentNode = nil
	sw.Trace = nil
	sw.Result = 0
	sw.stack = nil
}

// descend walks into a node. If it's a leaf, pause for input.
// If it's a composite, push a frame and descend into the first child.
func (sw *SimWalker) descend(node *model.NodeSpec, depth int) {
	sw.Trace = append(sw.Trace, SimStep{
		NodeName: node.Name,
		NodeType: node.Type,
		Depth:    depth,
		Event:    "enter",
	})

	if isLeaf(node) {
		sw.State = SimWaitingForInput
		sw.CurrentNode = node
		return
	}

	// Composite: push frame and descend into first child
	sw.stack = append(sw.stack, frameEntry{
		node:       node,
		childIndex: 0,
		results:    nil,
		depth:      depth,
	})
	if len(node.Children) > 0 {
		sw.descend(&node.Children[0], depth+1)
	}
}

// bubbleUp takes a child result and feeds it to the parent composite,
// continuing execution or completing the simulation.
func (sw *SimWalker) bubbleUp(childResult model.Status) {
	if len(sw.stack) == 0 {
		// No parent — this was the root (a single leaf tree)
		sw.State = SimComplete
		sw.Result = childResult
		return
	}

	frame := &sw.stack[len(sw.stack)-1]
	frame.results = append(frame.results, childResult)
	frame.childIndex++

	resolved, done := evaluateComposite(frame.node, frame.results, frame.childIndex)

	if done {
		// This composite is resolved
		sw.Trace = append(sw.Trace, SimStep{
			NodeName: frame.node.Name,
			NodeType: frame.node.Type,
			Status:   resolved,
			Depth:    frame.depth,
			Event:    "resolve",
		})

		// Pop frame and bubble further
		sw.stack = sw.stack[:len(sw.stack)-1]
		sw.bubbleUp(resolved)
		return
	}

	// Composite needs more children evaluated — descend into next child
	sw.descend(&frame.node.Children[frame.childIndex], frame.depth+1)
}

// evaluateComposite checks if a composite node can be resolved given
// the results so far. Returns (status, isDone).
func evaluateComposite(node *model.NodeSpec, results []model.Status, nextChild int) (model.Status, bool) {
	switch node.Type {
	case "sequence":
		return evalSequence(node, results, nextChild)
	case "selector":
		return evalSelector(node, results, nextChild)
	case "parallel":
		return evalParallel(node, results, nextChild)
	default:
		// Extension composites: treat like sequence by default
		return evalSequence(node, results, nextChild)
	}
}

func evalSequence(node *model.NodeSpec, results []model.Status, nextChild int) (model.Status, bool) {
	last := results[len(results)-1]

	// Sequence stops on first FAILURE or RUNNING
	if last == model.StatusFailure {
		return model.StatusFailure, true
	}
	if last == model.StatusRunning {
		return model.StatusRunning, true
	}

	// All children evaluated and all succeeded
	if nextChild >= len(node.Children) {
		return model.StatusSuccess, true
	}

	// Need more children
	return 0, false
}

func evalSelector(node *model.NodeSpec, results []model.Status, nextChild int) (model.Status, bool) {
	last := results[len(results)-1]

	// Selector stops on first SUCCESS or RUNNING
	if last == model.StatusSuccess {
		return model.StatusSuccess, true
	}
	if last == model.StatusRunning {
		return model.StatusRunning, true
	}

	// All children evaluated and all failed
	if nextChild >= len(node.Children) {
		return model.StatusFailure, true
	}

	// Need more children
	return 0, false
}

func evalParallel(node *model.NodeSpec, results []model.Status, nextChild int) (model.Status, bool) {
	// Parallel evaluates ALL children first
	if nextChild < len(node.Children) {
		return 0, false
	}

	// All children evaluated — default policy: any failure = failure
	for _, r := range results {
		if r == model.StatusFailure {
			return model.StatusFailure, true
		}
	}
	for _, r := range results {
		if r == model.StatusRunning {
			return model.StatusRunning, true
		}
	}
	return model.StatusSuccess, true
}

// applyDecorator transforms a leaf's result through its decorator (if any).
func applyDecorator(node *model.NodeSpec, status model.Status) model.Status {
	switch node.Decorator {
	case "negate":
		if status == model.StatusSuccess {
			return model.StatusFailure
		}
		if status == model.StatusFailure {
			return model.StatusSuccess
		}
		return status
	case "always_succeed":
		if status != model.StatusRunning {
			return model.StatusSuccess
		}
		return status
	case "always_fail":
		if status != model.StatusRunning {
			return model.StatusFailure
		}
		return status
	default:
		return status
	}
}

func isLeaf(node *model.NodeSpec) bool {
	return len(node.Children) == 0
}
