package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/treeedit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleSpec() *model.TreeSpec {
	return &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test-tree", Description: "A test tree"},
		Blackboard: []model.BlackboardVar{
			{Name: "target", Type: "Entity"},
			{Name: "health", Type: "float", Default: 100.0},
		},
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

// --- EditorModel Core ---

func TestNewEditorModel(t *testing.T) {
	spec := sampleSpec()
	em := NewEditorModel(spec, "/tmp/test.beetree.yaml")

	assert.Equal(t, spec, em.Spec)
	assert.Equal(t, "/tmp/test.beetree.yaml", em.FilePath)
	assert.Equal(t, ModeNavigate, em.Mode)
	assert.False(t, em.IsDirty())
	assert.Equal(t, "root", em.SelectedNodeName())
}

func TestNewEditorModel_DefaultSpec(t *testing.T) {
	em := NewEditorModel(nil, "")
	require.NotNil(t, em.Spec)
	assert.Equal(t, "selector", em.Spec.Tree.Type)
	assert.Equal(t, "root", em.Spec.Tree.Name)
}

func TestSelectedNode(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")

	node := em.SelectedNode()
	require.NotNil(t, node)
	assert.Equal(t, "root", node.Name)
	assert.Equal(t, "selector", node.Type)
}

func TestSelectNode(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")

	t.Run("select existing node", func(t *testing.T) {
		ok := em.SelectNode("combat")
		assert.True(t, ok)
		assert.Equal(t, "combat", em.SelectedNodeName())
	})

	t.Run("select non-existent node", func(t *testing.T) {
		ok := em.SelectNode("nonexistent")
		assert.False(t, ok)
		assert.Equal(t, "combat", em.SelectedNodeName()) // unchanged
	})

	t.Run("select leaf", func(t *testing.T) {
		ok := em.SelectNode("attack")
		assert.True(t, ok)
		assert.Equal(t, "attack", em.SelectedNodeName())
	})
}

func TestEditorMode(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")

	assert.Equal(t, ModeNavigate, em.Mode)

	em.Mode = ModeAddNode
	assert.Equal(t, ModeAddNode, em.Mode)

	em.Mode = ModeConfirmDelete
	assert.Equal(t, ModeConfirmDelete, em.Mode)
}

func TestDirtyFlag(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")

	assert.False(t, em.IsDirty())
	em.SetDirty(true)
	assert.True(t, em.IsDirty())
	em.SetDirty(false)
	assert.False(t, em.IsDirty())
}

func TestStatusMessage(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")

	assert.Equal(t, "", em.StatusMsg)
	em.StatusMsg = "Node added"
	assert.Equal(t, "Node added", em.StatusMsg)
}

// --- Navigation ---

func TestFlattenTree(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	flat := em.FlattenTree()

	// All nodes expanded by default: root, combat, has_target, attack, patrol
	require.Len(t, flat, 5)
	assert.Equal(t, "root", flat[0].Node.Name)
	assert.Equal(t, 0, flat[0].Depth)
	assert.Equal(t, "combat", flat[1].Node.Name)
	assert.Equal(t, 1, flat[1].Depth)
	assert.Equal(t, "has_target", flat[2].Node.Name)
	assert.Equal(t, 2, flat[2].Depth)
	assert.Equal(t, "attack", flat[3].Node.Name)
	assert.Equal(t, 2, flat[3].Depth)
	assert.Equal(t, "patrol", flat[4].Node.Name)
	assert.Equal(t, 1, flat[4].Depth)
}

func TestFlattenTree_Collapsed(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.CollapseNode("combat")

	flat := em.FlattenTree()
	// root, combat(collapsed), patrol
	require.Len(t, flat, 3)
	assert.Equal(t, "root", flat[0].Node.Name)
	assert.Equal(t, "combat", flat[1].Node.Name)
	assert.Equal(t, "patrol", flat[2].Node.Name)
}

func TestNavigateDown(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")

	em.NavigateDown()
	assert.Equal(t, "combat", em.SelectedNodeName())

	em.NavigateDown()
	assert.Equal(t, "has_target", em.SelectedNodeName())

	em.NavigateDown()
	assert.Equal(t, "attack", em.SelectedNodeName())

	em.NavigateDown()
	assert.Equal(t, "patrol", em.SelectedNodeName())

	// At bottom, stays on last
	em.NavigateDown()
	assert.Equal(t, "patrol", em.SelectedNodeName())
}

func TestNavigateUp(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("patrol")

	em.NavigateUp()
	assert.Equal(t, "attack", em.SelectedNodeName())

	em.NavigateUp()
	assert.Equal(t, "has_target", em.SelectedNodeName())

	em.NavigateUp()
	assert.Equal(t, "combat", em.SelectedNodeName())

	em.NavigateUp()
	assert.Equal(t, "root", em.SelectedNodeName())

	// At top, stays on first
	em.NavigateUp()
	assert.Equal(t, "root", em.SelectedNodeName())
}

func TestCollapseExpand(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")

	assert.True(t, em.IsExpanded("root"))
	assert.True(t, em.IsExpanded("combat"))

	em.CollapseNode("root")
	assert.False(t, em.IsExpanded("root"))

	em.ExpandNode("root")
	assert.True(t, em.IsExpanded("root"))
}

func TestToggleCollapse(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("combat")

	em.ToggleSelected()
	assert.False(t, em.IsExpanded("combat"))

	em.ToggleSelected()
	assert.True(t, em.IsExpanded("combat"))
}

// --- Node Operations ---

func TestAddNode(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("root")

	err := em.AddChild("new_action", "action", "NewAction")
	require.NoError(t, err)
	assert.True(t, em.IsDirty())

	// New node should exist in tree
	flat := em.FlattenTree()
	names := flatNames(flat)
	assert.Contains(t, names, "new_action")

	// Selection moves to new node
	assert.Equal(t, "new_action", em.SelectedNodeName())
}

func TestAddNode_ToLeafFails(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("attack")

	err := em.AddChild("child", "action", "")
	assert.Error(t, err)
	assert.False(t, em.IsDirty())
}

func TestAddNode_DuplicateNameFails(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("root")

	err := em.AddChild("patrol", "action", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDeleteNode(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("patrol")

	err := em.DeleteSelected()
	require.NoError(t, err)
	assert.True(t, em.IsDirty())

	flat := em.FlattenTree()
	names := flatNames(flat)
	assert.NotContains(t, names, "patrol")

	// Selection moves to previous sibling or parent
	assert.NotEqual(t, "patrol", em.SelectedNodeName())
}

func TestDeleteNode_Root(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("root")

	err := em.DeleteSelected()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "root")
}

func TestMoveNode(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")

	// Start move on "patrol"
	em.SelectNode("patrol")
	err := em.StartMove()
	require.NoError(t, err)
	assert.Equal(t, ModeMove, em.Mode)
	assert.Equal(t, "patrol", em.CutNodeName)

	// Complete move to "combat"
	em.SelectNode("combat")
	err = em.CompleteMove()
	require.NoError(t, err)
	assert.Equal(t, ModeNavigate, em.Mode)
	assert.True(t, em.IsDirty())

	// Patrol should now be a child of combat
	combat := findNode(&em.Spec.Tree, "combat")
	require.NotNil(t, combat)
	childNames := make([]string, len(combat.Children))
	for i, c := range combat.Children {
		childNames[i] = c.Name
	}
	assert.Contains(t, childNames, "patrol")
}

func TestMoveNode_ToLeafFails(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("patrol")
	em.StartMove()

	em.SelectNode("has_target")
	err := em.CompleteMove()
	assert.Error(t, err)
	assert.Equal(t, ModeMove, em.Mode) // stays in move mode
}

func TestCancelMove(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("patrol")
	em.StartMove()

	em.CancelMove()
	assert.Equal(t, ModeNavigate, em.Mode)
	assert.Empty(t, em.CutNodeName)
}

// --- Save ---

func TestSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.beetree.yaml")

	em := NewEditorModel(sampleSpec(), path)
	em.SetDirty(true)

	err := em.Save()
	require.NoError(t, err)
	assert.False(t, em.IsDirty())

	// File should exist
	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestSave_NoPath(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SetDirty(true)

	err := em.Save()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no file path")
	assert.True(t, em.IsDirty()) // still dirty
}

// --- Properties ---

func TestPropertiesForSelected(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("attack")

	props := em.PropertiesForSelected()
	assert.Equal(t, "attack", props.Name)
	assert.Equal(t, "action", props.Type)
	assert.Equal(t, "Attack", props.NodeClass)
}

func TestPropertiesForSelected_WithBlackboard(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")

	bb := em.BlackboardVars()
	require.Len(t, bb, 2)
	assert.Equal(t, "target", bb[0].Name)
	assert.Equal(t, "Entity", bb[0].Type)
	assert.Equal(t, "health", bb[1].Name)
}

// --- Simulation Integration ---

func TestEditorModel_StartSimulation(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.StartSimulation()

	assert.Equal(t, ModeSimulate, em.Mode)
	require.NotNil(t, em.Sim)
	assert.Equal(t, SimWaitingForInput, em.Sim.State)
	// Selection should move to the first leaf
	assert.Equal(t, "has_target", em.SelectedNodeName())
}

func TestEditorModel_SimResolve(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.StartSimulation()

	em.SimResolve(model.StatusSuccess) // has_target
	// Should advance to next leaf
	assert.Equal(t, "attack", em.SelectedNodeName())
	assert.Equal(t, SimWaitingForInput, em.Sim.State)

	em.SimResolve(model.StatusSuccess) // attack
	// Complete
	assert.Equal(t, SimComplete, em.Sim.State)
	assert.Equal(t, model.StatusSuccess, em.Sim.Result)
	assert.Contains(t, em.StatusMsg, "Success")
}

func TestEditorModel_SimResolveFailPath(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.StartSimulation()

	em.SimResolve(model.StatusFailure) // has_target fails → sequence fails
	// Selector falls through to patrol
	assert.Equal(t, "patrol", em.SelectedNodeName())

	em.SimResolve(model.StatusSuccess) // patrol succeeds
	assert.Equal(t, SimComplete, em.Sim.State)
	assert.Equal(t, model.StatusSuccess, em.Sim.Result)
}

func TestEditorModel_StopSimulation(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.StartSimulation()

	em.StopSimulation()
	assert.Equal(t, ModeNavigate, em.Mode)
	assert.Nil(t, em.Sim)
}

func TestEditorModel_SimNodeStatus(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.StartSimulation()

	em.SimResolve(model.StatusSuccess) // has_target

	status, found := em.SimNodeStatus("has_target")
	assert.True(t, found)
	assert.Equal(t, model.StatusSuccess, status)

	_, found = em.SimNodeStatus("patrol")
	assert.False(t, found) // not yet evaluated
}

func TestEditorModel_SimNodeStatus_NoSim(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	_, found := em.SimNodeStatus("root")
	assert.False(t, found)
}

// --- Undo ---

func TestUndo_AfterAdd(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("root")

	err := em.AddChild("new_node", "action", "NewAction")
	require.NoError(t, err)
	assert.True(t, em.CanUndo())

	ok := em.Undo()
	assert.True(t, ok)

	flat := em.FlattenTree()
	names := flatNames(flat)
	assert.NotContains(t, names, "new_node")
}

func TestUndo_AfterDelete(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("patrol")

	err := em.DeleteSelected()
	require.NoError(t, err)

	ok := em.Undo()
	assert.True(t, ok)

	flat := em.FlattenTree()
	names := flatNames(flat)
	assert.Contains(t, names, "patrol")
}

func TestUndo_AfterMove(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("patrol")
	require.NoError(t, em.StartMove())
	em.SelectNode("combat")
	require.NoError(t, em.CompleteMove())

	ok := em.Undo()
	assert.True(t, ok)

	// patrol should be back under root
	assert.Len(t, em.Spec.Tree.Children, 2)
	assert.Equal(t, "patrol", em.Spec.Tree.Children[1].Name)
}

func TestUndo_Empty(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	assert.False(t, em.CanUndo())
	assert.False(t, em.Undo())
}

func TestUndo_Multiple(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("root")

	em.AddChild("n1", "action", "")
	em.AddChild("n2", "action", "")

	em.Undo() // undo n2
	em.Undo() // undo n1

	flat := em.FlattenTree()
	names := flatNames(flat)
	assert.NotContains(t, names, "n1")
	assert.NotContains(t, names, "n2")
}

// --- Edit Node ---

func TestEditNode_Rename(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("patrol")

	err := em.EditNode(treeedit.NodeUpdates{Name: "guard"})
	require.NoError(t, err)
	assert.Equal(t, "guard", em.SelectedNodeName())
	assert.True(t, em.IsDirty())
}

func TestEditNode_ChangeType(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("patrol")

	newType := "condition"
	err := em.EditNode(treeedit.NodeUpdates{Type: &newType})
	require.NoError(t, err)

	node := em.SelectedNode()
	assert.Equal(t, "condition", node.Type)
}

func TestEditNode_UndoableRename(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("patrol")

	err := em.EditNode(treeedit.NodeUpdates{Name: "guard"})
	require.NoError(t, err)

	em.Undo()
	assert.Equal(t, "patrol", em.SelectedNodeName())
	node := findNode(&em.Spec.Tree, "patrol")
	require.NotNil(t, node)
}

func TestEditNode_DuplicateNameFails(t *testing.T) {
	em := NewEditorModel(sampleSpec(), "")
	em.SelectNode("patrol")

	err := em.EditNode(treeedit.NodeUpdates{Name: "combat"})
	assert.Error(t, err)
	assert.Equal(t, "patrol", em.SelectedNodeName()) // unchanged
}

// --- Save As ---

func TestSaveAs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new-tree.beetree.yaml")

	em := NewEditorModel(nil, "")
	em.SetDirty(true)

	err := em.SaveAs(path)
	require.NoError(t, err)
	assert.Equal(t, path, em.FilePath)
	assert.False(t, em.IsDirty())

	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestSaveAs_EmptyPath(t *testing.T) {
	em := NewEditorModel(nil, "")
	err := em.SaveAs("")
	assert.Error(t, err)
}

// --- Helpers ---

func flatNames(flat []FlatNode) []string {
	names := make([]string, len(flat))
	for i, fn := range flat {
		names[i] = fn.Node.Name
	}
	return names
}

func findNode(root *model.NodeSpec, name string) *model.NodeSpec {
	if root.Name == name {
		return root
	}
	for i := range root.Children {
		if found := findNode(&root.Children[i], name); found != nil {
			return found
		}
	}
	return nil
}
