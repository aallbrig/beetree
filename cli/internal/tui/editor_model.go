package tui

import (
	"fmt"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/treeedit"
	"gopkg.in/yaml.v3"
)

// EditorMode represents the current interaction mode of the editor.
type EditorMode int

const (
	ModeNavigate      EditorMode = iota
	ModeAddNode                         // type selector open
	ModeConfirmDelete                   // awaiting delete confirmation
	ModeMove                            // node cut, awaiting paste target
	ModeSimulate                        // interactive step-through simulation
	ModeEdit                            // property editor form open
	ModeSaveAs                          // save-as file path prompt
	ModeConfirmQuit                     // unsaved changes quit confirmation
)

const maxUndoDepth = 50

// FlatNode is a node in the flattened visible tree, used for cursor navigation.
type FlatNode struct {
	Node  *model.NodeSpec
	Depth int
}

// EditorModel manages all TUI state independent of the rendering framework.
type EditorModel struct {
	Spec        *model.TreeSpec
	FilePath    string
	Mode        EditorMode
	StatusMsg   string
	CutNodeName string // name of node being moved
	Sim         *SimWalker

	selectedName string
	dirty        bool
	expanded     map[string]bool
	undoStack    []undoSnapshot
}

// undoSnapshot stores tree state for undo.
type undoSnapshot struct {
	specYAML     []byte // serialized TreeSpec
	selectedName string
}

// NewEditorModel creates an EditorModel. If spec is nil, creates a default empty tree.
func NewEditorModel(spec *model.TreeSpec, filePath string) *EditorModel {
	if spec == nil {
		spec = &model.TreeSpec{
			Version:  "1.0",
			Metadata: model.Metadata{Name: "new-tree"},
			Tree: model.NodeSpec{
				Type: "selector",
				Name: "root",
			},
		}
	}

	em := &EditorModel{
		Spec:         spec,
		FilePath:     filePath,
		Mode:         ModeNavigate,
		selectedName: spec.Tree.Name,
		expanded:     make(map[string]bool),
	}

	// Expand all nodes by default
	em.expandAll(&spec.Tree)
	return em
}

func (em *EditorModel) expandAll(node *model.NodeSpec) {
	em.expanded[node.Name] = true
	for i := range node.Children {
		em.expandAll(&node.Children[i])
	}
}

// SelectedNodeName returns the name of the currently selected node.
func (em *EditorModel) SelectedNodeName() string {
	return em.selectedName
}

// SelectedNode returns a pointer to the currently selected NodeSpec.
func (em *EditorModel) SelectedNode() *model.NodeSpec {
	return treeedit.FindNode(&em.Spec.Tree, em.selectedName)
}

// SelectNode changes the selection. Returns false if name not found.
func (em *EditorModel) SelectNode(name string) bool {
	if treeedit.FindNode(&em.Spec.Tree, name) != nil {
		em.selectedName = name
		return true
	}
	return false
}

// IsDirty returns whether the tree has unsaved changes.
func (em *EditorModel) IsDirty() bool {
	return em.dirty
}

// SetDirty sets the dirty flag.
func (em *EditorModel) SetDirty(d bool) {
	em.dirty = d
}

// --- Navigation ---

// FlattenTree returns the tree as a flat list, respecting collapsed state.
func (em *EditorModel) FlattenTree() []FlatNode {
	var flat []FlatNode
	em.flattenNode(&em.Spec.Tree, 0, &flat)
	return flat
}

func (em *EditorModel) flattenNode(node *model.NodeSpec, depth int, flat *[]FlatNode) {
	*flat = append(*flat, FlatNode{Node: node, Depth: depth})
	if em.expanded[node.Name] {
		for i := range node.Children {
			em.flattenNode(&node.Children[i], depth+1, flat)
		}
	}
}

// NavigateDown moves selection to the next visible node.
func (em *EditorModel) NavigateDown() {
	flat := em.FlattenTree()
	for i, fn := range flat {
		if fn.Node.Name == em.selectedName && i+1 < len(flat) {
			em.selectedName = flat[i+1].Node.Name
			return
		}
	}
}

// NavigateUp moves selection to the previous visible node.
func (em *EditorModel) NavigateUp() {
	flat := em.FlattenTree()
	for i, fn := range flat {
		if fn.Node.Name == em.selectedName && i > 0 {
			em.selectedName = flat[i-1].Node.Name
			return
		}
	}
}

// IsExpanded returns whether a node's children are visible.
func (em *EditorModel) IsExpanded(name string) bool {
	return em.expanded[name]
}

// CollapseNode hides a node's children.
func (em *EditorModel) CollapseNode(name string) {
	em.expanded[name] = false
}

// ExpandNode shows a node's children.
func (em *EditorModel) ExpandNode(name string) {
	em.expanded[name] = true
}

// ToggleSelected collapses or expands the currently selected node.
func (em *EditorModel) ToggleSelected() {
	if em.expanded[em.selectedName] {
		em.CollapseNode(em.selectedName)
	} else {
		em.ExpandNode(em.selectedName)
	}
}

// --- Node Operations ---

// AddChild adds a new child node to the currently selected node.
func (em *EditorModel) AddChild(name, nodeType, nodeClass string) error {
	em.pushUndo()
	child := model.NodeSpec{
		Type: nodeType,
		Name: name,
		Node: nodeClass,
	}
	if err := treeedit.AddNode(&em.Spec.Tree, em.selectedName, child); err != nil {
		em.undoStack = em.undoStack[:len(em.undoStack)-1]
		return err
	}
	em.expanded[em.selectedName] = true
	em.expanded[name] = true
	em.selectedName = name
	em.dirty = true
	em.StatusMsg = fmt.Sprintf("Added %s node %q", nodeType, name)
	return nil
}

// DeleteSelected removes the currently selected node from the tree.
func (em *EditorModel) DeleteSelected() error {
	em.pushUndo()
	name := em.selectedName

	// Find a fallback selection before deleting
	flat := em.FlattenTree()
	fallback := em.Spec.Tree.Name
	for i, fn := range flat {
		if fn.Node.Name == name {
			if i > 0 {
				fallback = flat[i-1].Node.Name
			} else if i+1 < len(flat) {
				fallback = flat[i+1].Node.Name
			}
			break
		}
	}

	if err := treeedit.RemoveNode(&em.Spec.Tree, name); err != nil {
		em.undoStack = em.undoStack[:len(em.undoStack)-1]
		return err
	}
	delete(em.expanded, name)
	em.selectedName = fallback
	em.dirty = true
	em.StatusMsg = fmt.Sprintf("Deleted node %q", name)
	return nil
}

// StartMove begins a move operation on the currently selected node.
func (em *EditorModel) StartMove() error {
	if em.selectedName == em.Spec.Tree.Name {
		return fmt.Errorf("cannot move root node")
	}
	em.CutNodeName = em.selectedName
	em.Mode = ModeMove
	em.StatusMsg = fmt.Sprintf("Moving %q — select new parent, then press 'm' to place (Esc to cancel)", em.CutNodeName)
	return nil
}

// CompleteMove finishes the move, placing the cut node under the current selection.
func (em *EditorModel) CompleteMove() error {
	em.pushUndo()
	if err := treeedit.MoveNode(&em.Spec.Tree, em.CutNodeName, em.selectedName); err != nil {
		em.undoStack = em.undoStack[:len(em.undoStack)-1]
		em.StatusMsg = fmt.Sprintf("Move failed: %v", err)
		return err
	}
	em.expanded[em.selectedName] = true
	moved := em.CutNodeName
	em.CutNodeName = ""
	em.Mode = ModeNavigate
	em.dirty = true
	em.selectedName = moved
	em.StatusMsg = fmt.Sprintf("Moved %q under %q", moved, em.selectedName)
	return nil
}

// CancelMove cancels an in-progress move operation.
func (em *EditorModel) CancelMove() {
	em.CutNodeName = ""
	em.Mode = ModeNavigate
	em.StatusMsg = "Move cancelled"
}

// --- Save ---

// Save writes the spec to disk at FilePath.
func (em *EditorModel) Save() error {
	if em.FilePath == "" {
		return fmt.Errorf("no file path set")
	}
	if err := treeedit.SaveSpec(em.Spec, em.FilePath); err != nil {
		return err
	}
	em.dirty = false
	em.StatusMsg = fmt.Sprintf("Saved to %s", em.FilePath)
	return nil
}

// --- Undo ---

// pushUndo saves the current tree state to the undo stack.
func (em *EditorModel) pushUndo() {
	data, err := yaml.Marshal(em.Spec)
	if err != nil {
		return
	}
	em.undoStack = append(em.undoStack, undoSnapshot{
		specYAML:     data,
		selectedName: em.selectedName,
	})
	if len(em.undoStack) > maxUndoDepth {
		em.undoStack = em.undoStack[len(em.undoStack)-maxUndoDepth:]
	}
}

// Undo restores the previous tree state. Returns false if nothing to undo.
func (em *EditorModel) Undo() bool {
	if len(em.undoStack) == 0 {
		return false
	}
	snap := em.undoStack[len(em.undoStack)-1]
	em.undoStack = em.undoStack[:len(em.undoStack)-1]

	var spec model.TreeSpec
	if err := yaml.Unmarshal(snap.specYAML, &spec); err != nil {
		em.StatusMsg = fmt.Sprintf("Undo failed: %v", err)
		return false
	}
	em.Spec = &spec
	em.expanded = make(map[string]bool)
	em.expandAll(&em.Spec.Tree)
	if treeedit.FindNode(&em.Spec.Tree, snap.selectedName) != nil {
		em.selectedName = snap.selectedName
	} else {
		em.selectedName = em.Spec.Tree.Name
	}
	em.dirty = true
	em.StatusMsg = "Undone"
	return true
}

// CanUndo returns whether there are undo snapshots available.
func (em *EditorModel) CanUndo() bool {
	return len(em.undoStack) > 0
}

// --- Edit Node ---

// EditNode applies property changes to the currently selected node.
func (em *EditorModel) EditNode(updates treeedit.NodeUpdates) error {
	em.pushUndo()
	oldName := em.selectedName
	if err := treeedit.UpdateNode(&em.Spec.Tree, oldName, updates); err != nil {
		// Pop the undo snapshot since edit failed
		em.undoStack = em.undoStack[:len(em.undoStack)-1]
		return err
	}
	// If renamed, update selection and expanded state
	if updates.Name != "" && updates.Name != oldName {
		em.expanded[updates.Name] = em.expanded[oldName]
		delete(em.expanded, oldName)
		em.selectedName = updates.Name
	}
	em.dirty = true
	em.StatusMsg = fmt.Sprintf("Updated node %q", em.selectedName)
	return nil
}

// --- Save As ---

// SaveAs writes the spec to disk at the given path, updating FilePath.
func (em *EditorModel) SaveAs(path string) error {
	if path == "" {
		return fmt.Errorf("empty file path")
	}
	em.FilePath = path
	return em.Save()
}

// --- Properties ---

// PropertiesForSelected returns display properties for the selected node.
func (em *EditorModel) PropertiesForSelected() Properties {
	return NodeProperties(em.SelectedNode())
}

// BlackboardVars returns the spec's blackboard variables.
func (em *EditorModel) BlackboardVars() []model.BlackboardVar {
	return em.Spec.Blackboard
}

// --- Simulation ---

// StartSimulation enters interactive step-through simulation mode.
func (em *EditorModel) StartSimulation() {
	em.Sim = NewSimWalker(&em.Spec.Tree)
	em.Mode = ModeSimulate
	em.Sim.Step()
	if em.Sim.CurrentNode != nil {
		em.selectedName = em.Sim.CurrentNode.Name
	}
	em.StatusMsg = "Simulation started — choose result for each node"
}

// SimResolve provides the user's chosen status for the current leaf.
func (em *EditorModel) SimResolve(status model.Status) {
	if em.Sim == nil || em.Sim.State != SimWaitingForInput {
		return
	}
	em.Sim.Resolve(status)

	if em.Sim.State == SimComplete {
		em.StatusMsg = fmt.Sprintf("Simulation complete — tree returned %s", em.Sim.Result)
	} else if em.Sim.CurrentNode != nil {
		em.selectedName = em.Sim.CurrentNode.Name
		em.StatusMsg = fmt.Sprintf("Choose result for %q (%s)", em.Sim.CurrentNode.Name, em.Sim.CurrentNode.Type)
	}
}

// StopSimulation exits simulation mode.
func (em *EditorModel) StopSimulation() {
	em.Sim = nil
	em.Mode = ModeNavigate
	em.StatusMsg = "Simulation stopped"
}

// SimNodeStatus returns the resolved status for a node name if it has been
// evaluated during the current simulation. Returns (status, found).
func (em *EditorModel) SimNodeStatus(name string) (model.Status, bool) {
	if em.Sim == nil {
		return 0, false
	}
	// Walk trace backwards to find the most recent resolve for this node
	for i := len(em.Sim.Trace) - 1; i >= 0; i-- {
		s := em.Sim.Trace[i]
		if s.NodeName == name && s.Event == "resolve" {
			return s.Status, true
		}
	}
	return 0, false
}

