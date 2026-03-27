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

		template := fmt.Sprintf(`version: "1.0"
metadata:
  name: "%s"
  description: ""
  author: ""
  tags: []

blackboard: []

tree:
  type: "selector"
  name: "root"
  children:
    - type: "sequence"
      name: "main_behavior"
      children:
        - type: "condition"
          name: "check_precondition"
          node: "CheckPrecondition"
        - type: "action"
          name: "perform_action"
          node: "PerformAction"
    - type: "action"
      name: "fallback_action"
      node: "FallbackAction"
`, name)

		if err := os.WriteFile(path, []byte(template), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		cmd.Printf("✓ Created %s\n", path)
		cmd.Printf("  Edit the file to define your behavior tree\n")
		cmd.Printf("  Then run: beetree validate %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
