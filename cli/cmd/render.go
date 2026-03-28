package cmd

import (
	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/renderer"
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/spf13/cobra"
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render a .beetree.yaml spec as ASCII, Mermaid, or DOT diagram",
	Long: `Render a .beetree.yaml specification file as a visual tree diagram.

Supported formats:
  ascii   - Indented ASCII tree (default)
  mermaid - Mermaid flowchart syntax (GitHub-compatible)
  dot     - Graphviz DOT digraph`,
}

func makeRenderCmd(format, desc string, renderFn func(*model.TreeSpec) string) *cobra.Command {
	return &cobra.Command{
		Use:   format + " <file>",
		Short: "Render a .beetree.yaml spec as " + desc,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tree, err := spec.ParseFile(args[0])
			if err != nil {
				return err
			}
			cmd.Print(renderFn(tree))
			return nil
		},
	}
}

func init() {
	renderCmd.AddCommand(makeRenderCmd("ascii", "an ASCII tree", func(t *model.TreeSpec) string {
		return renderer.RenderSpecASCII(&t.Tree)
	}))
	renderCmd.AddCommand(makeRenderCmd("mermaid", "a Mermaid flowchart", func(t *model.TreeSpec) string {
		return renderer.RenderMermaid(&t.Tree)
	}))
	renderCmd.AddCommand(makeRenderCmd("dot", "a Graphviz DOT digraph", func(t *model.TreeSpec) string {
		return renderer.RenderDOT(&t.Tree)
	}))
	rootCmd.AddCommand(renderCmd)
}
