package cmd

import (
	"github.com/aallbrig/beetree-cli/internal/renderer"
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/spf13/cobra"
)

var renderFormat string

var renderCmd = &cobra.Command{
	Use:   "render <file>",
	Short: "Render a .beetree.yaml spec as ASCII or Mermaid diagram",
	Long: `Render a .beetree.yaml specification file as a visual tree diagram.

Supported formats:
  ascii   - Indented ASCII tree (default)
  mermaid - Mermaid flowchart syntax`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tree, err := spec.ParseFile(args[0])
		if err != nil {
			return err
		}

		var output string
		switch renderFormat {
		case "mermaid":
			output = renderer.RenderMermaid(&tree.Tree)
		default:
			output = renderer.RenderSpecASCII(&tree.Tree)
		}

		cmd.Print(output)
		return nil
	},
}

func init() {
	renderCmd.Flags().StringVar(&renderFormat, "format", "ascii", "Output format: ascii, mermaid")
	rootCmd.AddCommand(renderCmd)
}
