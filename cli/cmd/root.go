package cmd

import (
	"fmt"
	"os"

	"github.com/aallbrig/beetree-cli/internal/parser"
	"github.com/aallbrig/beetree-cli/internal/renderer"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "beetree",
	Short: "A behavior tree ecosystem CLI for authoring, validating, and generating engine code",
	Long: `BeeTree is a behavior tree ecosystem that makes it easy to author, share,
and deploy behavior trees to any game engine.

Define behavior trees in an engine-agnostic .beetree.yaml specification,
then generate native code for Unity (C#), Unreal (C++), and Godot (GDScript).`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		tree, err := parser.Parse(args[0])
		if err != nil {
			cmd.PrintErrf("Error parsing tree: %v\n", err)
			return
		}

		output, err := renderer.RenderASCII(tree)
		if err != nil {
			cmd.PrintErrf("Error rendering tree: %v\n", err)
			return
		}

		cmd.Println(output)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
