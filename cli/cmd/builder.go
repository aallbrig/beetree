package cmd

import (
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/aallbrig/beetree-cli/internal/tui"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

var builderCmd = &cobra.Command{
	Use:   "builder [file]",
	Short: "Interactive TUI editor for behavior trees",
	Long: `Launch the interactive behavior tree editor.

If a .beetree.yaml file is provided, it will be loaded for editing.
Otherwise, a new empty tree is created.

Key bindings:
  ↑/↓        Navigate tree
  ←/→        Collapse/expand
  a           Add child node
  e           Edit selected node properties
  d           Delete selected node
  m           Move node (press m again on target to place, Esc to cancel)
  u           Undo last change
  r           Run interactive simulation (step through tree, choose each result)
  s           Save to file (prompts for path if new)
  q           Quit (confirms if unsaved changes)

Simulation mode:
  S           Node returns Success
  F           Node returns Failure
  R           Node returns Running
  Esc         Stop simulation`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBuilder,
}

func init() {
	rootCmd.AddCommand(builderCmd)
}

func runBuilder(cmd *cobra.Command, args []string) error {
	var em *tui.EditorModel

	if len(args) == 1 {
		s, err := spec.ParseFile(args[0])
		if err != nil {
			return err
		}
		em = tui.NewEditorModel(s, args[0])
	} else {
		em = tui.NewEditorModel(nil, "")
	}

	app := tview.NewApplication()
	ev := tui.NewEditorView(app, em)
	if err := app.SetRoot(ev.Widget(), true).EnableMouse(true).Run(); err != nil {
		return err
	}
	return nil
}
