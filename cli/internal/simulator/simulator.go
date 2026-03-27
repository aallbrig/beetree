// Package simulator provides behavior tree simulation/dry-run capability.
// It walks a tree spec and simulates execution, producing a trace of
// node visits and their resulting statuses.
package simulator

import (
	"fmt"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/model"
)

// Status represents the result of a node execution.
type Status string

const (
	StatusSuccess Status = "SUCCESS"
	StatusFailure Status = "FAILURE"
	StatusRunning Status = "RUNNING"
)

// Step records a single event during simulation.
type Step struct {
	Depth    int
	NodeName string
	NodeType string
	Event    string // "enter", "execute", "exit"
	Status   Status
}

// Result holds the full simulation output.
type Result struct {
	NodeName string
	Status   Status
	Steps    []Step
}

// Simulate runs a dry-run of the behavior tree, returning a trace.
// Overrides map node names to forced statuses for testing scenarios.
func Simulate(spec *model.TreeSpec, overrides map[string]Status) *Result {
	sim := &simulator{
		overrides: overrides,
		steps:     nil,
	}
	status := sim.execute(&spec.Tree, 0)
	return &Result{
		NodeName: spec.Tree.Name,
		Status:   status,
		Steps:    sim.steps,
	}
}

type simulator struct {
	overrides map[string]Status
	steps     []Step
}

func (s *simulator) record(depth int, name, nodeType, event string, status Status) {
	s.steps = append(s.steps, Step{
		Depth:    depth,
		NodeName: name,
		NodeType: nodeType,
		Event:    event,
		Status:   status,
	})
}

func (s *simulator) execute(node *model.NodeSpec, depth int) Status {
	switch node.Type {
	case "action":
		return s.executeLeaf(node, depth)
	case "condition":
		return s.executeLeaf(node, depth)
	case "sequence":
		return s.executeSequence(node, depth)
	case "selector":
		return s.executeSelector(node, depth)
	case "parallel":
		return s.executeParallel(node, depth)
	case "decorator":
		return s.executeDecorator(node, depth)
	default:
		// Unknown type, treat as action
		return s.executeLeaf(node, depth)
	}
}

func (s *simulator) executeLeaf(node *model.NodeSpec, depth int) Status {
	status := StatusSuccess
	if s.overrides != nil {
		if override, ok := s.overrides[node.Name]; ok {
			status = override
		}
	}
	s.record(depth, node.Name, node.Type, "execute", status)
	return status
}

func (s *simulator) executeSequence(node *model.NodeSpec, depth int) Status {
	s.record(depth, node.Name, "sequence", "enter", "")
	for i := range node.Children {
		childStatus := s.execute(&node.Children[i], depth+1)
		if childStatus == StatusFailure {
			s.record(depth, node.Name, "sequence", "exit", StatusFailure)
			return StatusFailure
		}
		if childStatus == StatusRunning {
			s.record(depth, node.Name, "sequence", "exit", StatusRunning)
			return StatusRunning
		}
	}
	s.record(depth, node.Name, "sequence", "exit", StatusSuccess)
	return StatusSuccess
}

func (s *simulator) executeSelector(node *model.NodeSpec, depth int) Status {
	s.record(depth, node.Name, "selector", "enter", "")
	for i := range node.Children {
		childStatus := s.execute(&node.Children[i], depth+1)
		if childStatus == StatusSuccess {
			s.record(depth, node.Name, "selector", "exit", StatusSuccess)
			return StatusSuccess
		}
		if childStatus == StatusRunning {
			s.record(depth, node.Name, "selector", "exit", StatusRunning)
			return StatusRunning
		}
	}
	s.record(depth, node.Name, "selector", "exit", StatusFailure)
	return StatusFailure
}

func (s *simulator) executeParallel(node *model.NodeSpec, depth int) Status {
	s.record(depth, node.Name, "parallel", "enter", "")
	allSuccess := true
	anyRunning := false
	for i := range node.Children {
		childStatus := s.execute(&node.Children[i], depth+1)
		if childStatus == StatusFailure {
			s.record(depth, node.Name, "parallel", "exit", StatusFailure)
			return StatusFailure
		}
		if childStatus == StatusRunning {
			anyRunning = true
			allSuccess = false
		}
	}
	if anyRunning {
		s.record(depth, node.Name, "parallel", "exit", StatusRunning)
		return StatusRunning
	}
	if allSuccess {
		s.record(depth, node.Name, "parallel", "exit", StatusSuccess)
		return StatusSuccess
	}
	s.record(depth, node.Name, "parallel", "exit", StatusSuccess)
	return StatusSuccess
}

func (s *simulator) executeDecorator(node *model.NodeSpec, depth int) Status {
	s.record(depth, node.Name, "decorator:"+node.Decorator, "enter", "")
	if len(node.Children) == 0 {
		s.record(depth, node.Name, "decorator:"+node.Decorator, "exit", StatusFailure)
		return StatusFailure
	}
	childStatus := s.execute(&node.Children[0], depth+1)
	result := applyDecorator(node.Decorator, childStatus)
	s.record(depth, node.Name, "decorator:"+node.Decorator, "exit", result)
	return result
}

func applyDecorator(decorator string, childStatus Status) Status {
	switch decorator {
	case "negate":
		if childStatus == StatusSuccess {
			return StatusFailure
		}
		if childStatus == StatusFailure {
			return StatusSuccess
		}
		return childStatus
	case "always_succeed":
		return StatusSuccess
	case "always_fail":
		return StatusFailure
	default:
		return childStatus
	}
}

// FormatTrace produces a human-readable trace from a simulation result.
func FormatTrace(result *Result) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Simulation: %s → %s\n", result.NodeName, result.Status))
	b.WriteString(strings.Repeat("─", 50) + "\n")
	for _, step := range result.Steps {
		indent := strings.Repeat("  ", step.Depth)
		statusStr := ""
		if step.Status != "" {
			statusStr = fmt.Sprintf(" → %s", step.Status)
		}
		symbol := eventSymbol(step.Event)
		b.WriteString(fmt.Sprintf("%s%s [%s] %s%s\n", indent, symbol, step.NodeType, step.NodeName, statusStr))
	}
	return b.String()
}

func eventSymbol(event string) string {
	switch event {
	case "enter":
		return "▶"
	case "exit":
		return "◀"
	case "execute":
		return "●"
	default:
		return "·"
	}
}
