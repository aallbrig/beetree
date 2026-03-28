package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new behavior tree specification",
	Long: `Creates a new .beetree.yaml file in the trees/ directory
with a starter template.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		filename := name + ".beetree.yaml"
		path := filepath.Join("trees", filename)

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists", path)
		}

		if err := os.MkdirAll("trees", 0755); err != nil {
			return fmt.Errorf("creating trees directory: %w", err)
		}

		template := fmt.Sprintf(`# ─────────────────────────────────────────────────────────────
# Behavior Tree: %s
# ─────────────────────────────────────────────────────────────
# Edit this file to define your AI behavior, then run:
#   beetree validate trees/%s.beetree.yaml
#   beetree render   trees/%s.beetree.yaml
#   beetree simulate trees/%s.beetree.yaml
#
# TIP: See examples/ in the BeeTree repo for real-world patterns.
#
# QUICK REFERENCE:
#   Selector  — tries children until one SUCCEEDS (like if/else)
#   Sequence  — runs children in order, ALL must succeed (like AND)
#   Action    — does something (move, attack, wait)
#   Condition — checks something (has target? low health?)
#   Parallel  — runs all children at the same time
#   Decorator — wraps a child (repeat, negate, timeout)
# ─────────────────────────────────────────────────────────────

version: "1.0"
metadata:
  name: %s
  description: ""
  author: ""
  tags: []

# Blackboard: shared variables that nodes read and write.
# This is how nodes communicate — one node writes "target",
# another reads it to decide what to do.
blackboard:
  - name: target
    type: Entity
    description: "Current target (null if none)"
  - name: health
    type: float
    default: 100.0
    description: "Current health points"

# The behavior tree. Read top-to-bottom:
# The SELECTOR tries each child in order — first success wins.
tree:
  type: selector
  name: root
  children:
    # Priority 1: main behavior (sequence = ALL steps must succeed)
    - type: sequence
      name: main_behavior
      children:
        # Condition: check if we should act
        - type: condition
          name: check_precondition
          node: CheckPrecondition
          description: "Does the precondition hold?"
        # Action: do the thing
        - type: action
          name: perform_action
          node: PerformAction
          description: "Execute the main behavior"

    # Priority 2: fallback (always reached if main_behavior fails)
    - type: action
      name: fallback_action
      node: FallbackAction
      description: "Default behavior when nothing else applies"
`, name, name, name, name, name)

		if err := os.WriteFile(path, []byte(template), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		cmd.Printf("✓ Created %s\n\n", path)
		cmd.Printf("  Next steps:\n")
		cmd.Printf("    beetree render %s         # visualize the tree\n", path)
		cmd.Printf("    beetree simulate %s       # dry-run execution\n", path)
		cmd.Printf("    beetree node add %s root my_node --type action --node MyNode\n", path)
		cmd.Printf("    beetree generate unity %s  # generate engine code\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
