package cmd

import (
	"fmt"

	"github.com/aallbrig/beetree-cli/internal/differ"
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <file1> <file2>",
	Short: "Compare two behavior tree specs",
	Long:  "Show structural differences between two .beetree.yaml specs.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		specA, err := spec.ParseFile(args[0])
		if err != nil {
			return fmt.Errorf("parse %s: %w", args[0], err)
		}
		specB, err := spec.ParseFile(args[1])
		if err != nil {
			return fmt.Errorf("parse %s: %w", args[1], err)
		}

		changes := differ.Diff(specA, specB)
		cmd.Print(differ.FormatDiff(changes))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
