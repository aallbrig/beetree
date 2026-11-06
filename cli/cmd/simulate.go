package cmd

import (
	"fmt"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/simulator"
	"github.com/spf13/cobra"
)

var (
	simulateOverrides []string
	simulateStep      bool
)

var simulateCmd = &cobra.Command{
	Use:   "simulate <file>",
	Short: "Simulate behavior tree execution (dry-run)",
	Long: `Walk the behavior tree and simulate execution, showing node traversal
and resulting statuses. Use --override to force specific node outcomes, or
--step to pause at each leaf and choose Success, Failure or Running yourself
(answer s, f or r); composites resolve from your answers and the trace prints
at the end. --step uses the same walker as the TUI.

Example:
  beetree simulate trees/enemy-ai.beetree.yaml
  beetree simulate trees/enemy-ai.beetree.yaml --override check_health=FAILURE
  beetree simulate trees/enemy-ai.beetree.yaml --step`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tree, err := loadValidSpec(args[0])
		if err != nil {
			return err
		}

		if simulateStep {
			if len(simulateOverrides) > 0 {
				return fmt.Errorf("--step and --override cannot be combined")
			}
			return runStepThrough(&tree.Tree, cmd.InOrStdin(), cmd.OutOrStdout())
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
	simulateCmd.Flags().BoolVar(&simulateStep, "step", false, "Pause at each leaf and choose its status interactively")
	rootCmd.AddCommand(simulateCmd)
}
