package cmd

import (
	"fmt"

	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/aallbrig/beetree-cli/internal/validator"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <file>...",
	Short: "Validate one or more .beetree.yaml specification files",
	Example: `  beetree validate trees/enemy-ai.beetree.yaml
  beetree validate examples/*.beetree.yaml`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return validateOne(cmd, args[0])
		}
		failed := 0
		for _, path := range args {
			if err := validateOne(cmd, path); err != nil {
				failed++
				cmd.PrintErrf("%v\n", err)
			}
		}
		if failed > 0 {
			return fmt.Errorf("%d of %d files failed validation", failed, len(args))
		}
		return nil
	},
}

func validateOne(cmd *cobra.Command, path string) error {
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
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
