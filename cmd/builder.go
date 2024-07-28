package cmd

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

type TreeNode struct {
	Name     string
	Children []*TreeNode
}

func (n *TreeNode) AddChild(child *TreeNode) {
	n.Children = append(n.Children, child)
}

var builderCmd = &cobra.Command{
	Use:   "builder",
	Short: "Interactive builder for creating behaviors",
	Long:  `This command launches an interactive REPL for creating and managing behaviors.`,
	Run:   runBuilder,
}

func init() {
	rootCmd.AddCommand(builderCmd)
}

func runBuilder(cmd *cobra.Command, args []string) {
	app := tview.NewApplication()

	rootNode := &TreeNode{Name: "Root"}
	treeView := tview.NewTreeView()

	root := tview.NewTreeNode(rootNode.Name).SetReference(rootNode)
	treeView.SetRoot(root).SetCurrentNode(root)

	updateTreeView := func(node *TreeNode, tNode *tview.TreeNode) {
		for _, child := range node.Children {
			childNode := tview.NewTreeNode(child.Name).SetReference(child)
			tNode.AddChild(childNode)
		}
	}

	treeView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			selectedNode := treeView.GetCurrentNode()
			ref := selectedNode.GetReference().(*TreeNode)
			childName := fmt.Sprintf("Child of %s", ref.Name)
			childNode := &TreeNode{Name: childName}
			ref.AddChild(childNode)
			updateTreeView(ref, selectedNode)
			app.Draw()
		}
		return event
	})

	if err := app.SetRoot(treeView, true).Run(); err != nil {
		panic(err)
	}
}
