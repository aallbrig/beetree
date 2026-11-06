package tui

import (
	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/walker"
)

// The step-through walker lives in internal/walker so non-TUI callers (the
// API server) can use it without linking tview. These aliases keep the TUI
// vocabulary unchanged.
type (
	SimState  = walker.SimState
	SimStep   = walker.SimStep
	SimWalker = walker.SimWalker
)

const (
	SimReady           = walker.SimReady
	SimWaitingForInput = walker.SimWaitingForInput
	SimComplete        = walker.SimComplete
)

// NewSimWalker creates a walker for the given tree root.
func NewSimWalker(root *model.NodeSpec) *SimWalker { return walker.NewSimWalker(root) }
