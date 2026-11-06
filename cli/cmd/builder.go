package cmd

import (
	"fmt"
	"strings"

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
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBuilder,
}

// keyHelp renders tui.KeyBindings grouped by mode so --help cannot drift.
func keyHelp() string {
	var b strings.Builder
	mode := ""
	for _, kb := range tui.KeyBindings {
		if kb.Mode != mode {
			mode = kb.Mode
			fmt.Fprintf(&b, "\n%s mode:\n", mode)
		}
		fmt.Fprintf(&b, "  %-11s %s\n", kb.Key, kb.Action)
	}
	return b.String()
}

func init() {
	builderCmd.Long = strings.TrimRight(builderCmd.Long, "\n") + "\n" + keyHelp()
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
