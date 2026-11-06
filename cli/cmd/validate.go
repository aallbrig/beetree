package cmd

import (
	"fmt"

	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/aallbrig/beetree-cli/internal/validator"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a .beetree.yaml specification file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		treeSpec, err := spec.ParseFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		errs := validator.Validate(treeSpec)
		if len(errs) > 0 {
			cmd.PrintErrf("Validation errors in %s:\n", path)
			for _, e := range errs {
				cmd.PrintErrf("  - %s\n", e.Error())
			}
			return fmt.Errorf("found %d validation error(s)", len(errs))
		}

		cmd.Printf("✓ %s is valid (%d nodes)\n", path, spec.NodeCount(&treeSpec.Tree))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
