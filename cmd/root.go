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
	Short: "A CLI tool to generate ASCII art representations of behavior trees",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please provide a behavior tree string")
			os.Exit(1)
		}

		tree, err := parser.Parse(args[0])
		if err != nil {
			fmt.Printf("Error parsing tree: %v\n", err)
			os.Exit(1)
		}

		output, err := renderer.RenderASCII(tree)
		if err != nil {
			fmt.Printf("Error rendering tree: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(output)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
