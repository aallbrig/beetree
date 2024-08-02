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
	rootNode.AddChild(&TreeNode{Name: "Condition"})
	seqNode := &TreeNode{Name: "Sequence"}
	seqNode.AddChild(&TreeNode{Name: "Condition"})
	seqNode.AddChild(&TreeNode{Name: "Task1"})
	seqNode.AddChild(&TreeNode{Name: "Task2"})
	rootNode.AddChild(seqNode)
	rootNode.AddChild(&TreeNode{Name: "Task3"})
	treeView := tview.NewTreeView()

	root := tview.NewTreeNode(rootNode.Name).SetReference(rootNode)
	treeView.SetRoot(root).SetCurrentNode(root)
	var populateTreeView func(*TreeNode, *tview.TreeNode)
	populateTreeView = func(node *TreeNode, tNode *tview.TreeNode) {
		for _, child := range node.Children {
			childNode := tview.NewTreeNode(child.Name).SetReference(child)
			tNode.AddChild(childNode)
			populateTreeView(child, childNode)
		}
	}
	populateTreeView(rootNode, root)

	// Add a text view for debugging
	debugView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	treeView.SetFocusFunc(func() {
		debugView.SetText(fmt.Sprintf("Focused tree node: %s", treeView.GetCurrentNode().GetReference().(*TreeNode).Name))
	})
	treeView.SetSelectedFunc(func(node *tview.TreeNode) {
		debugView.SetText(fmt.Sprintf("Selected tree node: %s", node.GetReference().(*TreeNode).Name))
	})

	// Create a flex layout
	flex := tview.NewFlex().
		AddItem(treeView, 0, 1, true).
		AddItem(debugView, 0, 1, false)

	treeView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		debugView.SetText(fmt.Sprintf("Key pressed: %v", event.Key()))

		if event.Key() == tcell.KeyEnter {
			selectedNode := treeView.GetCurrentNode()
			ref := selectedNode.GetReference().(*TreeNode)
			childName := fmt.Sprintf("Child of %s", ref.Name)
			childNode := &TreeNode{Name: childName}
			ref.AddChild(childNode)

			newTreeNode := tview.NewTreeNode(childName).SetReference(childNode)
			selectedNode.AddChild(newTreeNode)

			selectedNode.Expand()
			treeView.SetCurrentNode(newTreeNode)

			debugView.SetText(debugView.GetText(true) + "\nAdded new node: " + childName)

			return nil
		}
		return event
	})

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
