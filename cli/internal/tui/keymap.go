package tui

// KeyBinding documents one TUI key. The docs page at website/content/docs/tui.md
// lists every entry; a test keeps the page, this map and the in-app help in step.
type KeyBinding struct {
	Mode   string
	Key    string
	Action string
}

// KeyBindings is the full key map, grouped by mode.
var KeyBindings = []KeyBinding{
	{"Navigate", "a", "Add a child node"},
	{"Navigate", "e", "Edit the selected node"},
	{"Navigate", "d", "Delete the selected node"},
	{"Navigate", "c", "Duplicate the selected node"},
	{"Navigate", "m", "Start a move, then press m on the target to place it"},
	{"Navigate", "p", "Edit node parameters"},
	{"Navigate", "b", "Edit blackboard variables"},
	{"Navigate", "u", "Undo the last change"},
	{"Navigate", "U", "Redo"},
	{"Navigate", "r", "Run the step-through simulation"},
	{"Navigate", "v", "Validate the tree"},
	{"Navigate", "/", "Search nodes by name"},
	{"Navigate", "s", "Save"},
	{"Navigate", "S", "Save as a new path"},
	{"Navigate", "q", "Quit (prompts if unsaved)"},
	{"Navigate", "?", "Show the help overlay"},
	{"Navigate", "Up/Down", "Move the selection"},
	{"Navigate", "Left/Right", "Collapse or expand the node"},
	{"Move", "Up/Down", "Choose the target parent"},
	{"Move", "m", "Place the node under the target"},
	{"Move", "Esc", "Cancel the move"},
	{"Simulate", "S", "Resolve the paused leaf as Success"},
	{"Simulate", "F", "Resolve the paused leaf as Failure"},
	{"Simulate", "R", "Resolve the paused leaf as Running"},
	{"Simulate", "Esc", "Stop the simulation"},
	{"Simulate", "q", "Leave the finished simulation"},
}
