package tree_editor

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/looplab/fsm"
	"github.com/rivo/tview"
)

const (
	state_navigation = "navigation"
)

type Editor struct {
	TreeBeingBuilt *TreeNode
	Widget         tview.Primitive
	treeView       *tview.TreeView
	debugView      *tview.TextView
	fsm            *fsm.FSM
}

func NewEditor(app *tview.Application) *Editor {
	e := &Editor{
		fsm: fsm.NewFSM(
			state_navigation,
			fsm.Events{
				{Name: state_navigation, Src: []string{state_navigation}, Dst: state_navigation},
			},
			fsm.Callbacks{},
		),
	}
	e.treeView = tview.NewTreeView()
	seqNode := &TreeNode{Name: "Sequence"}
	seqNode.AddChild(&TreeNode{Name: "Condition"})
	seqNode.AddChild(&TreeNode{Name: "Task1"})
	seqNode.AddChild(&TreeNode{Name: "Task2"})
	e.TreeBeingBuilt = &TreeNode{Name: "Root"}
	e.TreeBeingBuilt.AddChild(&TreeNode{Name: "Condition"})
	e.TreeBeingBuilt.AddChild(seqNode)
	e.TreeBeingBuilt.AddChild(&TreeNode{Name: "Task3"})
	root := tview.NewTreeNode(e.TreeBeingBuilt.Name).SetReference(e.TreeBeingBuilt)
	e.debugView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	e.treeView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		e.debugView.SetText(fmt.Sprintf("Key pressed: %v", event.Key()))

		if event.Key() == tcell.KeyEnter {
			selectedNode := e.treeView.GetCurrentNode()
			ref := selectedNode.GetReference().(*TreeNode)
			childName := fmt.Sprintf("Child of %s", ref.Name)
			childNode := &TreeNode{Name: childName}
			ref.AddChild(childNode)

			newTreeNode := tview.NewTreeNode(childName).SetReference(childNode)
			selectedNode.AddChild(newTreeNode)

			selectedNode.Expand()
			e.treeView.SetCurrentNode(newTreeNode)

			e.debugView.SetText(e.debugView.GetText(true) + "\nAdded new node: " + childName)

			return nil
		}
		return event
	})

	e.treeView.SetRoot(root).SetCurrentNode(root)
	var populateTreeView func(*TreeNode, *tview.TreeNode)
	populateTreeView = func(node *TreeNode, tNode *tview.TreeNode) {
		for _, child := range node.Children {
			childNode := tview.NewTreeNode(child.Name).SetReference(child)
			tNode.AddChild(childNode)
			populateTreeView(child, childNode)
		}
	}
	populateTreeView(e.TreeBeingBuilt, root)

	e.treeView.SetFocusFunc(func() {
		e.debugView.SetText(fmt.Sprintf("Focused tree node: %s", e.treeView.GetCurrentNode().GetReference().(*TreeNode).Name))
	})
	e.treeView.SetSelectedFunc(func(node *tview.TreeNode) {
		e.debugView.SetText(fmt.Sprintf("Selected tree node: %s", node.GetReference().(*TreeNode).Name))
	})
	e.Widget = tview.NewFlex().
		AddItem(e.treeView, 0, 1, true).
		AddItem(e.debugView, 0, 1, false)

	return e
}

type TreeNode struct {
	Name     string
	Children []*TreeNode
}

func (n *TreeNode) AddChild(child *TreeNode) {
	n.Children = append(n.Children, child)
}
