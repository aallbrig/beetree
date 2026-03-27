package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initName string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new BeeTree project",
	Long: `Creates a beetree.yaml manifest and project directory structure
in the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := initName
		if name == "" {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			name = filepath.Base(dir)
		}

		manifestPath := "beetree.yaml"
		if _, err := os.Stat(manifestPath); err == nil {
			return fmt.Errorf("beetree.yaml already exists in this directory")
		}

		dirs := []string{"trees", "subtrees"}
		for _, d := range dirs {
			if err := os.MkdirAll(d, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", d, err)
			}
		}

		manifest := fmt.Sprintf(`version: "1.0"
metadata:
  name: "%s"
  description: ""
  author: ""
  tags: []
`, name)

		if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
			return fmt.Errorf("writing beetree.yaml: %w", err)
		}

		cmd.Printf("✓ Initialized BeeTree project %q\n", name)
		cmd.Printf("  Created: beetree.yaml, trees/, subtrees/\n")
		cmd.Printf("  Next: beetree new <tree-name> to create a behavior tree\n")
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initName, "name", "", "Project name (defaults to directory name)")
	rootCmd.AddCommand(initCmd)
}
