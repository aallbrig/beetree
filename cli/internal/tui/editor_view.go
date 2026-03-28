package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/treeedit"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// EditorView wires the EditorModel to tview widgets for the 3-pane TUI layout.
type EditorView struct {
	App        *tview.Application
	Model      *EditorModel
	Root       *tview.Flex
	treeView   *tview.TreeView
	propsView  *tview.TextView
	statusView *tview.TextView
	addModal   *tview.List
	nameInput  *tview.InputField
	editForm   *tview.Form
	saveAsInput *tview.InputField
	quitModal  *tview.List
	helpView   *tview.TextView
	pages      *tview.Pages

	pendingAddType string
}

// NewEditorView creates the full TUI layout from an EditorModel.
func NewEditorView(app *tview.Application, em *EditorModel) *EditorView {
	ev := &EditorView{
		App:   app,
		Model: em,
	}
	ev.buildLayout()
	ev.syncTreeView()
	ev.syncPropsView()
	ev.syncStatusBar()
	return ev
}

func (ev *EditorView) buildLayout() {
	// Tree pane (left)
	ev.treeView = tview.NewTreeView()
	ev.treeView.SetBorder(true).SetTitle(" Tree View ")
	ev.treeView.SetSelectedFunc(func(node *tview.TreeNode) {
		ref, ok := node.GetReference().(string)
		if !ok {
			return
		}
		ev.Model.SelectNode(ref)
		ev.Model.ToggleSelected()
		ev.syncTreeView()
		ev.syncPropsView()
	})

	// Properties pane (right)
	ev.propsView = tview.NewTextView()
	ev.propsView.SetDynamicColors(true).SetBorder(true).SetTitle(" Properties ")

	// Status bar (bottom)
	ev.statusView = tview.NewTextView()
	ev.statusView.SetDynamicColors(true).SetBorder(true).SetTitle(" Commands ")
	ev.statusView.SetTextAlign(tview.AlignLeft)

	// Add-node modal
	ev.addModal = tview.NewList().ShowSecondaryText(false)
	ev.addModal.SetBorder(true).SetTitle(" Select Node Type ")

	ev.nameInput = tview.NewInputField()
	ev.nameInput.SetLabel("Node name: ").SetFieldWidth(30)
	ev.nameInput.SetBorder(true).SetTitle(" Name New Node ")

	// Edit form (populated dynamically when opened)
	ev.editForm = tview.NewForm()
	ev.editForm.SetBorder(true).SetTitle(" Edit Node ")

	// Save-as input
	ev.saveAsInput = tview.NewInputField()
	ev.saveAsInput.SetLabel("Save as: ").SetFieldWidth(40)
	ev.saveAsInput.SetBorder(true).SetTitle(" Save As ")

	// Quit confirmation
	ev.quitModal = tview.NewList().ShowSecondaryText(false)
	ev.quitModal.SetBorder(true).SetTitle(" Unsaved Changes ")

	// Help overlay
	ev.helpView = tview.NewTextView()
	ev.helpView.SetDynamicColors(true).SetScrollable(true)
	ev.helpView.SetBorder(true).SetTitle(" Help — press Esc or ? to close ")
	ev.helpView.SetText(helpText)
	ev.helpView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || (event.Key() == tcell.KeyRune && event.Rune() == '?') {
			ev.pages.HidePage("help")
			ev.App.SetFocus(ev.treeView)
			return nil
		}
		return event
	})

	// Main layout
	topPane := tview.NewFlex().
		AddItem(ev.treeView, 0, 2, true).
		AddItem(ev.propsView, 0, 1, false)

	ev.Root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topPane, 0, 1, true).
		AddItem(ev.statusView, 3, 0, false)

	// Pages for modal overlays
	ev.pages = tview.NewPages().
		AddPage("main", ev.Root, true, true).
		AddPage("add-type", makeModal(ev.addModal, 40, 15), true, false).
		AddPage("add-name", makeModal(ev.nameInput, 50, 3), true, false).
		AddPage("edit-node", makeModal(ev.editForm, 55, 14), true, false).
		AddPage("save-as", makeModal(ev.saveAsInput, 60, 3), true, false).
		AddPage("quit-confirm", makeModal(ev.quitModal, 40, 5), true, false).
		AddPage("help", makeModal(ev.helpView, 72, 30), true, false)

	ev.setupKeyBindings()
}

func makeModal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height+2, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)
}

func (ev *EditorView) setupKeyBindings() {
	ev.treeView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch ev.Model.Mode {
		case ModeNavigate:
			return ev.handleNavigateKey(event)
		case ModeMove:
			return ev.handleMoveKey(event)
		case ModeSimulate:
			return ev.handleSimulateKey(event)
		default:
			return event
		}
	})
}

func (ev *EditorView) handleNavigateKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'a':
			ev.openAddNodeModal()
			return nil
		case 'e':
			ev.openEditForm()
			return nil
		case 'd':
			ev.doDelete()
			return nil
		case 'm':
			ev.doStartMove()
			return nil
		case 's':
			ev.doSave()
			return nil
		case 'u':
			ev.doUndo()
			return nil
		case 'r':
			ev.doStartSimulation()
			return nil
		case 'q':
			ev.doQuit()
			return nil
		case '?':
			ev.showHelp()
			return nil
		}
	case tcell.KeyLeft:
		if ev.Model.IsExpanded(ev.Model.SelectedNodeName()) {
			ev.Model.CollapseNode(ev.Model.SelectedNodeName())
			ev.syncTreeView()
		}
		return nil
	case tcell.KeyRight:
		if !ev.Model.IsExpanded(ev.Model.SelectedNodeName()) {
			ev.Model.ExpandNode(ev.Model.SelectedNodeName())
			ev.syncTreeView()
		}
		return nil
	case tcell.KeyUp:
		ev.Model.NavigateUp()
		ev.syncTreeView()
		ev.syncPropsView()
		return nil
	case tcell.KeyDown:
		ev.Model.NavigateDown()
		ev.syncTreeView()
		ev.syncPropsView()
		return nil
	}
	return event
}

func (ev *EditorView) handleMoveKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		if event.Rune() == 'm' {
			if err := ev.Model.CompleteMove(); err != nil {
				ev.Model.StatusMsg = fmt.Sprintf("Move failed: %v", err)
			}
			ev.syncAll()
			return nil
		}
	case tcell.KeyEscape:
		ev.Model.CancelMove()
		ev.syncAll()
		return nil
	case tcell.KeyUp:
		ev.Model.NavigateUp()
		ev.syncTreeView()
		ev.syncPropsView()
		return nil
	case tcell.KeyDown:
		ev.Model.NavigateDown()
		ev.syncTreeView()
		ev.syncPropsView()
		return nil
	}
	return event
}

func (ev *EditorView) handleSimulateKey(event *tcell.EventKey) *tcell.EventKey {
	if ev.Model.Sim == nil {
		return event
	}

	switch event.Key() {
	case tcell.KeyEscape:
		ev.Model.StopSimulation()
		ev.syncAll()
		return nil
	case tcell.KeyRune:
		if ev.Model.Sim.State == SimWaitingForInput {
			switch event.Rune() {
			case 's', 'S':
				ev.Model.SimResolve(model.StatusSuccess)
				ev.syncAll()
				return nil
			case 'f', 'F':
				ev.Model.SimResolve(model.StatusFailure)
				ev.syncAll()
				return nil
			case 'r', 'R':
				ev.Model.SimResolve(model.StatusRunning)
				ev.syncAll()
				return nil
			}
		}
		if ev.Model.Sim.State == SimComplete {
			switch event.Rune() {
			case 'q', 'Q':
				ev.Model.StopSimulation()
				ev.syncAll()
				return nil
			}
		}
	}
	return event
}

func (ev *EditorView) doStartSimulation() {
	ev.Model.StartSimulation()
	ev.syncAll()
}

func (ev *EditorView) openAddNodeModal() {
	ev.Model.Mode = ModeAddNode
	ev.addModal.Clear()

	types := AvailableNodeTypes()
	sort.Slice(types, func(i, j int) bool {
		if types[i].Category != types[j].Category {
			return types[i].Category < types[j].Category
		}
		return types[i].Name < types[j].Name
	})

	for _, nt := range types {
		entry := nt
		ev.addModal.AddItem(fmt.Sprintf("[%s] %s", entry.Category, entry.Name), "", 0, func() {
			ev.pendingAddType = entry.Name
			ev.pages.HidePage("add-type")
			ev.openNameInput()
		})
	}

	ev.addModal.SetDoneFunc(func() {
		ev.Model.Mode = ModeNavigate
		ev.pages.HidePage("add-type")
		ev.App.SetFocus(ev.treeView)
		ev.syncStatusBar()
	})

	ev.pages.ShowPage("add-type")
	ev.App.SetFocus(ev.addModal)
}

func (ev *EditorView) openNameInput() {
	ev.nameInput.SetText("")
	ev.nameInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			name := ev.nameInput.GetText()
			if name != "" {
				if err := ev.Model.AddChild(name, ev.pendingAddType, ""); err != nil {
					ev.Model.StatusMsg = fmt.Sprintf("Error: %v", err)
				}
			}
		}
		ev.Model.Mode = ModeNavigate
		ev.pages.HidePage("add-name")
		ev.App.SetFocus(ev.treeView)
		ev.syncAll()
	})

	ev.pages.ShowPage("add-name")
	ev.App.SetFocus(ev.nameInput)
}

func (ev *EditorView) openEditForm() {
	node := ev.Model.SelectedNode()
	if node == nil {
		return
	}
	ev.Model.Mode = ModeEdit
	ev.editForm.Clear(true)

	// Pre-populate fields with current values
	nameField := tview.NewInputField().SetLabel("Name: ").SetText(node.Name).SetFieldWidth(30)
	ev.editForm.AddFormItem(nameField)

	// Type dropdown
	typeOptions := []string{"selector", "sequence", "parallel", "action", "condition", "decorator",
		"utility_selector", "active_selector", "random_selector", "random_sequence", "subtree"}
	typeIdx := 0
	for i, t := range typeOptions {
		if t == node.Type {
			typeIdx = i
			break
		}
	}
	ev.editForm.AddDropDown("Type: ", typeOptions, typeIdx, nil)

	ev.editForm.AddInputField("Node Class: ", node.Node, 30, nil, nil)
	ev.editForm.AddInputField("Decorator: ", node.Decorator, 30, nil, nil)

	ev.editForm.AddButton("Save", func() {
		newName := ev.editForm.GetFormItemByLabel("Name: ").(*tview.InputField).GetText()
		_, newType := ev.editForm.GetFormItemByLabel("Type: ").(*tview.DropDown).GetCurrentOption()
		newClass := ev.editForm.GetFormItemByLabel("Node Class: ").(*tview.InputField).GetText()
		newDecorator := ev.editForm.GetFormItemByLabel("Decorator: ").(*tview.InputField).GetText()

		updates := treeedit.NodeUpdates{
			Name:      newName,
			Type:      &newType,
			NodeClass: &newClass,
			Decorator: &newDecorator,
		}
		if err := ev.Model.EditNode(updates); err != nil {
			ev.Model.StatusMsg = fmt.Sprintf("Edit failed: %v", err)
		}
		ev.closeEditForm()
	})
	ev.editForm.AddButton("Cancel", func() {
		ev.closeEditForm()
	})

	ev.pages.ShowPage("edit-node")
	ev.App.SetFocus(ev.editForm)
}

func (ev *EditorView) closeEditForm() {
	ev.Model.Mode = ModeNavigate
	ev.pages.HidePage("edit-node")
	ev.App.SetFocus(ev.treeView)
	ev.syncAll()
}

func (ev *EditorView) doUndo() {
	if ev.Model.Undo() {
		ev.syncAll()
	} else {
		ev.Model.StatusMsg = "Nothing to undo"
		ev.syncStatusBar()
	}
}

func (ev *EditorView) doQuit() {
	if !ev.Model.IsDirty() {
		ev.App.Stop()
		return
	}
	ev.Model.Mode = ModeConfirmQuit
	ev.quitModal.Clear()
	ev.quitModal.AddItem("[s] Save and quit", "", 's', func() {
		ev.doSave()
		ev.App.Stop()
	})
	ev.quitModal.AddItem("[q] Quit without saving", "", 'q', func() {
		ev.App.Stop()
	})
	ev.quitModal.AddItem("[Esc] Cancel", "", 0, func() {
		ev.closeQuitModal()
	})
	ev.quitModal.SetDoneFunc(func() {
		ev.closeQuitModal()
	})
	ev.pages.ShowPage("quit-confirm")
	ev.App.SetFocus(ev.quitModal)
}

func (ev *EditorView) closeQuitModal() {
	ev.Model.Mode = ModeNavigate
	ev.pages.HidePage("quit-confirm")
	ev.App.SetFocus(ev.treeView)
	ev.syncStatusBar()
}

func (ev *EditorView) openSaveAsModal() {
	ev.Model.Mode = ModeSaveAs
	ev.saveAsInput.SetText("tree.beetree.yaml")
	ev.saveAsInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			path := ev.saveAsInput.GetText()
			if path != "" {
				if err := ev.Model.SaveAs(path); err != nil {
					ev.Model.StatusMsg = fmt.Sprintf("Save failed: %v", err)
				}
			}
		}
		ev.Model.Mode = ModeNavigate
		ev.pages.HidePage("save-as")
		ev.App.SetFocus(ev.treeView)
		ev.syncAll()
	})
	ev.pages.ShowPage("save-as")
	ev.App.SetFocus(ev.saveAsInput)
}

func (ev *EditorView) doDelete() {
	if err := ev.Model.DeleteSelected(); err != nil {
		ev.Model.StatusMsg = fmt.Sprintf("Error: %v", err)
	}
	ev.syncAll()
}

func (ev *EditorView) doStartMove() {
	if err := ev.Model.StartMove(); err != nil {
		ev.Model.StatusMsg = fmt.Sprintf("Error: %v", err)
	}
	ev.syncAll()
}

func (ev *EditorView) doSave() {
	if ev.Model.FilePath == "" {
		ev.openSaveAsModal()
		return
	}
	if err := ev.Model.Save(); err != nil {
		ev.Model.StatusMsg = fmt.Sprintf("Error: %v", err)
	}
	ev.syncAll()
}

func (ev *EditorView) syncAll() {
	ev.syncTreeView()
	ev.syncPropsView()
	ev.syncStatusBar()
}

func (ev *EditorView) syncTreeView() {
	flat := ev.Model.FlattenTree()
	if len(flat) == 0 {
		return
	}

	// Build tview tree from flat list using a stack
	root := tview.NewTreeNode(ev.nodeDisplayText(flat[0].Node)).
		SetReference(flat[0].Node.Name).
		SetExpanded(ev.Model.IsExpanded(flat[0].Node.Name)).
		SetColor(ev.nodeColor(flat[0].Node.Name))

	type stackEntry struct {
		tNode *tview.TreeNode
		depth int
	}
	stack := []stackEntry{{root, 0}}

	for _, fn := range flat[1:] {
		tNode := tview.NewTreeNode(ev.nodeDisplayText(fn.Node)).
			SetReference(fn.Node.Name).
			SetExpanded(ev.Model.IsExpanded(fn.Node.Name)).
			SetColor(ev.nodeColor(fn.Node.Name))

		// Find the parent at the correct depth
		for len(stack) > 0 && stack[len(stack)-1].depth >= fn.Depth {
			stack = stack[:len(stack)-1]
		}
		if len(stack) > 0 {
			stack[len(stack)-1].tNode.AddChild(tNode)
		}
		stack = append(stack, stackEntry{tNode, fn.Depth})
	}

	ev.treeView.SetRoot(root).SetCurrentNode(findTreeNode(root, ev.Model.SelectedNodeName()))
}

// nodeColor returns the display color for a node based on editor state.
func (ev *EditorView) nodeColor(name string) tcell.Color {
	// Simulation colors take priority
	if ev.Model.Mode == ModeSimulate && ev.Model.Sim != nil {
		// Current node waiting for input: bright cyan
		if ev.Model.Sim.State == SimWaitingForInput && ev.Model.Sim.CurrentNode != nil && ev.Model.Sim.CurrentNode.Name == name {
			return tcell.ColorAqua
		}
		// Already resolved: color by status
		if status, found := ev.Model.SimNodeStatus(name); found {
			switch status {
			case model.StatusSuccess:
				return tcell.ColorGreen
			case model.StatusFailure:
				return tcell.ColorRed
			case model.StatusRunning:
				return tcell.ColorYellow
			}
		}
	}

	if name == ev.Model.SelectedNodeName() {
		return tcell.ColorGreen
	}
	if name == ev.Model.CutNodeName {
		return tcell.ColorYellow
	}
	return tcell.ColorWhite
}

// nodeDisplayText returns the label for a node, appending status during simulation.
func (ev *EditorView) nodeDisplayText(node *model.NodeSpec) string {
	label := NodeLabel(node, ev.Model.Spec.Notation)
	if ev.Model.Mode == ModeSimulate {
		if status, found := ev.Model.SimNodeStatus(node.Name); found {
			switch status {
			case model.StatusSuccess:
				label += " ✓"
			case model.StatusFailure:
				label += " ✗"
			case model.StatusRunning:
				label += " ⟳"
			}
		}
		if ev.Model.Sim != nil && ev.Model.Sim.State == SimWaitingForInput && ev.Model.Sim.CurrentNode != nil && ev.Model.Sim.CurrentNode.Name == node.Name {
			label += " ← YOU DECIDE"
		}
	}
	return label
}

func findTreeNode(root *tview.TreeNode, name string) *tview.TreeNode {
	if ref, ok := root.GetReference().(string); ok && ref == name {
		return root
	}
	for _, child := range root.GetChildren() {
		if found := findTreeNode(child, name); found != nil {
			return found
		}
	}
	return root
}

func (ev *EditorView) syncPropsView() {
	// During simulation, show trace instead of properties
	if ev.Model.Mode == ModeSimulate && ev.Model.Sim != nil {
		ev.syncSimTraceView()
		return
	}

	props := ev.Model.PropertiesForSelected()
	var b strings.Builder

	b.WriteString(fmt.Sprintf("[yellow]Name:[-] %s\n", props.Name))
	b.WriteString(fmt.Sprintf("[yellow]Type:[-] %s %s\n", NodeTypeTag(props.Type), props.Type))
	b.WriteString(fmt.Sprintf("[yellow]Children:[-] %d\n", props.ChildCount))

	if props.NodeClass != "" {
		b.WriteString(fmt.Sprintf("[yellow]Node Class:[-] %s\n", props.NodeClass))
	}
	if props.Decorator != "" {
		b.WriteString(fmt.Sprintf("[yellow]Decorator:[-] %s\n", props.Decorator))
	}
	if props.SubtreeRef != "" {
		b.WriteString(fmt.Sprintf("[yellow]Subtree:[-] %s\n", props.SubtreeRef))
	}
	if props.SubtreeFile != "" {
		b.WriteString(fmt.Sprintf("[yellow]File:[-] %s\n", props.SubtreeFile))
	}

	if len(props.Parameters) > 0 {
		b.WriteString("\n[green]── Parameters ──[-]\n")
		for k, v := range props.Parameters {
			b.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}

	bb := ev.Model.BlackboardVars()
	if len(bb) > 0 {
		b.WriteString("\n[green]── Blackboard ──[-]\n")
		for _, v := range bb {
			defVal := "null"
			if v.Default != nil {
				defVal = fmt.Sprintf("%v", v.Default)
			}
			b.WriteString(fmt.Sprintf("  %s: %s (%s)\n", v.Name, v.Type, defVal))
		}
	}

	ev.propsView.SetText(b.String())
}

func (ev *EditorView) syncSimTraceView() {
	var b strings.Builder
	b.WriteString("[yellow::b]── Simulation Trace ──[-:-:-]\n\n")

	for _, step := range ev.Model.Sim.Trace {
		indent := strings.Repeat("  ", step.Depth)
		switch step.Event {
		case "enter":
			b.WriteString(fmt.Sprintf("%s[blue]▶[-] %s %s\n", indent, NodeTypeTag(step.NodeType), step.NodeName))
		case "resolve":
			var color, symbol string
			switch step.Status {
			case model.StatusSuccess:
				color, symbol = "green", "✓"
			case model.StatusFailure:
				color, symbol = "red", "✗"
			case model.StatusRunning:
				color, symbol = "yellow", "⟳"
			}
			b.WriteString(fmt.Sprintf("%s[%s]%s %s → %s[-]\n", indent, color, symbol, step.NodeName, step.Status))
		}
	}

	if ev.Model.Sim.State == SimComplete {
		b.WriteString(fmt.Sprintf("\n[white::b]══ Result: %s ══[-:-:-]\n", ev.Model.Sim.Result))
	}

	ev.propsView.SetText(b.String())
}

func (ev *EditorView) syncStatusBar() {
	dirty := ""
	if ev.Model.IsDirty() {
		dirty = " [red][MODIFIED][-]"
	}

	var status string
	switch ev.Model.Mode {
	case ModeNavigate:
		status = "[a]dd  [e]dit  [d]elete  [m]ove  [u]ndo  [r]un sim  [s]ave  [q]uit  [?]help"
	case ModeMove:
		status = fmt.Sprintf("[yellow]MOVE MODE[-] — navigate to target, [m] to place, [Esc] cancel  (moving: %s)", ev.Model.CutNodeName)
	case ModeAddNode:
		status = "[↑↓] select type  [Enter] confirm  [Esc] cancel"
	case ModeSimulate:
		if ev.Model.Sim != nil && ev.Model.Sim.State == SimWaitingForInput {
			status = "[aqua]SIM[-] [green][S][-]uccess  [red][F][-]ailure  [yellow][R][-]unning  │  [Esc] stop"
		} else if ev.Model.Sim != nil && ev.Model.Sim.State == SimComplete {
			status = fmt.Sprintf("[aqua]SIM COMPLETE[-] — tree returned [::b]%s[-:-:-]  │  [Q]uit sim  [Esc] stop", ev.Model.Sim.Result)
		} else {
			status = "[aqua]SIM[-] [Esc] stop"
		}
	default:
		status = ""
	}

	msg := ""
	if ev.Model.StatusMsg != "" {
		msg = fmt.Sprintf("  │  %s", ev.Model.StatusMsg)
	}

	file := ""
	if ev.Model.FilePath != "" {
		file = fmt.Sprintf("  │  %s", ev.Model.FilePath)
	}

	ev.statusView.SetText(fmt.Sprintf(" %s%s%s%s", status, msg, file, dirty))
}

// Widget returns the root primitive for use with tview.Application.SetRoot.
func (ev *EditorView) Widget() tview.Primitive {
	return ev.pages
}

func (ev *EditorView) showHelp() {
	ev.pages.ShowPage("help")
	ev.App.SetFocus(ev.helpView)
}

const helpText = `[yellow::b]═══ Key Bindings ═══[-::-]

[green::b]Navigate Mode[-::-]
  [aqua]a[-]         Add a child node
  [aqua]e[-]         Edit selected node
  [aqua]d[-]         Delete selected node
  [aqua]m[-]         Start move (cut), then navigate + [aqua]m[-] to paste
  [aqua]u[-]         Undo last change
  [aqua]r[-]         Run interactive simulation
  [aqua]s[-]         Save file
  [aqua]S[-]         Save as (new path)
  [aqua]q[-]         Quit (prompts if unsaved)
  [aqua]?[-]         Show this help
  [aqua]↑ ↓[-]       Navigate tree
  [aqua]← →[-]       Collapse / expand node

[green::b]Move Mode[-::-]
  [aqua]↑ ↓[-]       Navigate to target parent
  [aqua]m[-]         Place node under target
  [aqua]Esc[-]       Cancel move

[green::b]Simulation Mode[-::-]
  [aqua]S[-]         Resolve leaf as [green]Success[-]
  [aqua]F[-]         Resolve leaf as [red]Failure[-]
  [aqua]R[-]         Resolve leaf as [yellow]Running[-]
  [aqua]Esc[-]       Stop simulation

[yellow::b]═══ Behavior Tree Concepts ═══[-::-]

[green::b]Composite Nodes[-::-] — have one or more children
  [white::b]Sequence[-::-]    Runs children left-to-right. Succeeds if ALL
               succeed. Fails on first failure. (AND logic)
  [white::b]Selector[-::-]    Runs children left-to-right. Succeeds on first
               success. Fails if ALL fail. (OR logic)
  [white::b]Parallel[-::-]    Runs all children simultaneously. Outcome
               depends on success/failure policy.

[green::b]Leaf Nodes[-::-] — perform actual work
  [white::b]Action[-::-]      Does something (move, attack, play animation).
               Returns Success, Failure, or Running.
  [white::b]Condition[-::-]   Checks something (has target? health > 50?).
               Returns Success or Failure.

[green::b]Decorators[-::-] — wrap a single child, modify its behavior
  [white::b]Negate[-::-]      Inverts child result (Success↔Failure)
  [white::b]Repeat[-::-]      Re-runs child N times or until failure
  [white::b]AlwaysSucceed[-::-] / [white::b]AlwaysFail[-::-]  Forces a result
  [white::b]UntilFail[-::-]   Repeats child until it fails
  [white::b]Timeout[-::-]     Fails child after duration
  [white::b]Cooldown[-::-]    Blocks re-entry for a duration
  [white::b]Retry[-::-]       Re-runs child on failure, up to N times

[green::b]Blackboard[-::-] — shared key-value store for tree state
  Nodes read/write variables (e.g., "target", "health").
  Defined in the spec with types and optional defaults.

[dim]Scroll with ↑↓. Press Esc or ? to close.[-]`
