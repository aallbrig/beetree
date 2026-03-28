package cmd

import (
	"github.com/aallbrig/beetree-cli/internal/renderer"
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/spf13/cobra"
)

var renderFormat string

var renderCmd = &cobra.Command{
	Use:   "render <file>",
	Short: "Render a .beetree.yaml spec as a tree diagram",
	Long: `Render a .beetree.yaml specification file as a visual tree diagram.

Supported formats:
  ascii   - Indented ASCII tree with [TAG] labels (default)
  sigil   - Unicode sigil tree (→ ? ! ¿ ◇ ⇒)
  compact - Sigil tree with indentation only (ideal for diffing)
  oneline - Single-line S-expression (for commit messages, search)
  mermaid - Mermaid flowchart syntax
  dot     - Graphviz DOT digraph`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tree, err := spec.ParseFile(args[0])
		if err != nil {
			return err
		}

		var output string
		switch renderFormat {
		case "sigil":
			output = renderer.RenderSigil(&tree.Tree, tree.Notation)
		case "compact":
			output = renderer.RenderCompact(&tree.Tree, tree.Notation)
		case "oneline":
			output = renderer.RenderOneline(&tree.Tree, tree.Notation)
		case "mermaid":
			output = renderer.RenderMermaid(&tree.Tree)
		case "dot":
			output = renderer.RenderDOT(&tree.Tree)
		default:
			output = renderer.RenderSpecASCII(&tree.Tree)
		}

		cmd.Print(output)
		return nil
	},
}

func init() {
	renderCmd.Flags().StringVar(&renderFormat, "format", "ascii", "Output format: ascii, sigil, compact, oneline, mermaid, dot")
	rootCmd.AddCommand(renderCmd)
}
