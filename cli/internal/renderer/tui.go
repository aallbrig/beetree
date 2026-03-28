package renderer

import (
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
