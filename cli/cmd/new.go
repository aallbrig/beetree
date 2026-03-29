package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var newTemplate string

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new behavior tree specification",
	Long: `Creates a new .beetree.yaml file in the trees/ directory.

Available templates (--template):
  default   Starter tree with selector, sequence, condition, and action
  patrol    NPC patrol loop with waypoints
  combat    Priority-based combat AI with flee/attack/chase/idle
  utility   Utility-selector based decision making
  blank     Minimal empty tree (just a root selector)`,
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

		tmpl, ok := treeTemplates[newTemplate]
		if !ok {
			return fmt.Errorf("unknown template %q (available: default, patrol, combat, utility, blank)", newTemplate)
		}

		content := tmpl(name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		cmd.Printf("✓ Created %s (template: %s)\n\n", path, newTemplate)
		cmd.Printf("  Next steps:\n")
		cmd.Printf("    beetree render %s         # visualize the tree\n", path)
		cmd.Printf("    beetree simulate %s       # dry-run execution\n", path)
		cmd.Printf("    beetree node add %s root my_node --type action --node MyNode\n", path)
		cmd.Printf("    beetree generate unity %s  # generate engine code\n", path)
		return nil
	},
}

func init() {
	newCmd.Flags().StringVar(&newTemplate, "template", "default", "tree template (default, patrol, combat, utility, blank)")
	rootCmd.AddCommand(newCmd)
}

var treeTemplates = map[string]func(name string) string{
	"default": templateDefault,
	"patrol":  templatePatrol,
	"combat":  templateCombat,
	"utility": templateUtility,
	"blank":   templateBlank,
}

func templateHeader(name string) string {
	return fmt.Sprintf(`# ─────────────────────────────────────────────────────────────
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
#   Selector  → tries children until one SUCCEEDS (like if/else)
#   Sequence  → runs children in order, ALL must succeed (like AND)
#   Action    ! does something (move, attack, wait)
#   Condition ¿ checks something (has target? low health?)
#   Parallel  ⇒ runs all children at the same time
#   Decorator ◇ wraps a child (repeat, negate, timeout)
# ─────────────────────────────────────────────────────────────
`, name, name, name, name)
}

func templateDefault(name string) string {
	return templateHeader(name) + fmt.Sprintf(`
version: "1.0"
metadata:
  name: %s
  description: ""
  author: ""
  tags: []

blackboard:
  - name: target
    type: Entity
    description: "Current target (null if none)"
  - name: health
    type: float
    default: 100.0
    description: "Current health points"

tree:
  type: selector
  name: root
  children:
    - type: sequence
      name: main_behavior
      children:
        - type: condition
          name: check_precondition
          node: CheckPrecondition
        - type: action
          name: perform_action
          node: PerformAction

    - type: action
      name: fallback_action
      node: FallbackAction
`, name)
}

func templatePatrol(name string) string {
	return templateHeader(name) + fmt.Sprintf(`
version: "1.0"
metadata:
  name: %s
  description: "NPC patrols between waypoints in a loop"
  author: ""
  tags: [patrol, npc, movement]

blackboard:
  - name: waypoints
    type: Vector3[]
    description: "List of patrol waypoints"
  - name: waypoint_index
    type: int
    default: 0
    description: "Current waypoint index"
  - name: wait_time
    type: float
    default: 2.0
    description: "Seconds to wait at each waypoint"
  - name: current_waypoint
    type: Vector3
    description: "Current destination waypoint"

tree:
  type: decorator
  name: patrol_loop
  decorator: repeat
  children:
    - type: sequence
      name: patrol_cycle
      children:
        - type: action
          name: move_to_waypoint
          node: MoveToWaypoint
          description: "Walk to the current waypoint"
        - type: action
          name: wait_at_waypoint
          node: WaitAtWaypoint
          description: "Pause at the waypoint"
        - type: action
          name: next_waypoint
          node: NextWaypoint
          description: "Advance to the next waypoint index"
`, name)
}

func templateCombat(name string) string {
	return templateHeader(name) + fmt.Sprintf(`
version: "1.0"
metadata:
  name: %s
  description: "Priority-based combat AI with flee, attack, chase, and idle"
  author: ""
  tags: [combat, ai, enemy]

blackboard:
  - name: health
    type: float
    default: 100.0
    description: "Current health points"
  - name: health_threshold
    type: float
    default: 20.0
    description: "Health below which NPC flees"
  - name: target
    type: Entity
    description: "Current combat target"
  - name: attack_range
    type: float
    default: 5.0
    description: "Range within which attack is possible"

tree:
  type: selector
  name: combat_root
  children:
    # Highest priority: flee when health is critical
    - type: sequence
      name: flee_behavior
      children:
        - type: condition
          name: is_health_low
          node: IsHealthLow
          description: "Is health below threshold?"
        - type: action
          name: flee
          node: Flee
          description: "Run away from combat"

    # Attack if target is in range
    - type: sequence
      name: attack_behavior
      children:
        - type: condition
          name: has_target
          node: HasTarget
        - type: condition
          name: is_in_range
          node: IsInRange
          description: "Is target within attack range?"
        - type: action
          name: attack
          node: Attack

    # Chase if target exists but out of range
    - type: sequence
      name: chase_behavior
      children:
        - type: condition
          name: has_target_chase
          node: HasTarget
        - type: action
          name: chase
          node: ChaseTarget
          description: "Move toward the target"

    # Fallback: idle
    - type: action
      name: idle
      node: Idle
      description: "Stand idle when nothing else to do"
`, name)
}

func templateUtility(name string) string {
	return templateHeader(name) + fmt.Sprintf(`
version: "1.0"
metadata:
  name: %s
  description: "Utility-based AI that scores actions by priority"
  author: ""
  tags: [utility, ai, decision-making]

blackboard:
  - name: hunger
    type: float
    default: 0.0
    description: "Hunger level (0-100)"
  - name: fatigue
    type: float
    default: 0.0
    description: "Fatigue level (0-100)"
  - name: danger
    type: float
    default: 0.0
    description: "Perceived danger level (0-100)"

# utility_selector picks the child with the highest utility score.
# Each child should write its score to the blackboard before evaluation.
tree:
  type: utility_selector
  name: decide
  children:
    - type: sequence
      name: eat_behavior
      children:
        - type: condition
          name: is_hungry
          node: IsHungry
          description: "Hunger above threshold?"
        - type: action
          name: find_food
          node: FindFood
        - type: action
          name: eat
          node: Eat

    - type: sequence
      name: rest_behavior
      children:
        - type: condition
          name: is_tired
          node: IsTired
          description: "Fatigue above threshold?"
        - type: action
          name: find_shelter
          node: FindShelter
        - type: action
          name: rest
          node: Rest

    - type: sequence
      name: flee_danger
      children:
        - type: condition
          name: is_danger_high
          node: IsDangerHigh
        - type: action
          name: flee_to_safety
          node: FleeToSafety

    - type: action
      name: wander
      node: Wander
      description: "Default: wander randomly"
`, name)
}

func templateBlank(name string) string {
	return fmt.Sprintf(`version: "1.0"
metadata:
  name: %s
  description: ""

tree:
  type: selector
  name: root
  children:
    - type: action
      name: my_action
      node: MyAction
`, name)
}
