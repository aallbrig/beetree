package cmd

import (
	"fmt"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/simulator"
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/spf13/cobra"
)

var (
	simulateOverrides []string
)

var simulateCmd = &cobra.Command{
	Use:   "simulate <file>",
	Short: "Simulate behavior tree execution (dry-run)",
	Long: `Walk the behavior tree and simulate execution, showing node traversal
and resulting statuses. Use --override to force specific node outcomes.

Example:
  beetree simulate trees/enemy-ai.beetree.yaml
  beetree simulate trees/enemy-ai.beetree.yaml --override check_health=FAILURE`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tree, err := spec.ParseFile(args[0])
		if err != nil {
			return fmt.Errorf("parse spec: %w", err)
		}

		overrides, err := parseOverrides(simulateOverrides)
		if err != nil {
			return err
		}

		result := simulator.Simulate(tree, overrides)
		cmd.Print(simulator.FormatTrace(result))
		return nil
	},
}

func parseOverrides(raw []string) (map[string]simulator.Status, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	overrides := make(map[string]simulator.Status)
	for _, o := range raw {
		parts := strings.SplitN(o, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid override %q: expected name=STATUS", o)
		}
		status := simulator.Status(strings.ToUpper(parts[1]))
		switch status {
		case simulator.StatusSuccess, simulator.StatusFailure, simulator.StatusRunning:
			overrides[parts[0]] = status
		default:
			return nil, fmt.Errorf("invalid status %q: must be SUCCESS, FAILURE, or RUNNING", parts[1])
		}
	}
	return overrides, nil
}

func init() {
	simulateCmd.Flags().StringSliceVar(&simulateOverrides, "override", nil, "Force node outcome: name=STATUS (can repeat)")
	rootCmd.AddCommand(simulateCmd)
}
