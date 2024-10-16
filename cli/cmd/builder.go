package cmd

import (
	"github.com/aallbrig/beetree-cli/internal/tree_editor"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

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
	treeEditor := tree_editor.NewEditor(app)
	if err := app.SetRoot(treeEditor.Widget, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
