package renderer

import (
	"fmt"
	"github.com/aallbrig/beetree-cli/internal/tree"
	"github.com/rivo/tview"
)

func RenderTree(root *tree.Node, tviewRoot *tview.TreeNode) {
	if root == nil {
		return
	}

	for _, child := range root.Children {
		tviewChild := tview.NewTreeNode("")
		RenderTree(child, tviewChild)
		tviewRoot.AddChild(tviewChild)
	}
}

func getBehaviorDescription(b tree.Behavior) string {
	if b == nil {
		return "Empty"
	}

	switch behavior := b.(type) {
	case *tree.Task:
		return fmt.Sprintf("Task: %s", behavior.Name)
	case *tree.Condition:
		return fmt.Sprintf("Condition: %s", behavior.Name)
	case *tree.Decorator:
		return fmt.Sprintf("Decorator: %s", behavior.Name)
	case *tree.Sequence:
		return fmt.Sprintf("Sequence: %s", behavior.Name)
	case *tree.Fallback:
		return fmt.Sprintf("Fallback: %s", behavior.Name)
	case *tree.Parallel:
		return fmt.Sprintf("Parellel: %s", behavior.Name)
	default:
		return "Unknown Behavior"
	}
}
