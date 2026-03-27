package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var registryPullCmd = &cobra.Command{
	Use:   "pull <owner/name>",
	Short: "Download a behavior tree spec from the registry",
	Long:  "Pull a published .beetree.yaml spec to the local trees/ directory.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getRegistryClient()
		data, err := client.Pull(context.Background(), args[0])
		if err != nil {
			return fmt.Errorf("pull failed: %w", err)
		}

		treesDir := "trees"
		if err := os.MkdirAll(treesDir, 0755); err != nil {
			return fmt.Errorf("create trees directory: %w", err)
		}

		// Use the name part for the local filename
		name := filepath.Base(args[0])
		outPath := filepath.Join(treesDir, name+".beetree.yaml")
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return fmt.Errorf("write file: %w", err)
		}

		cmd.Printf("✓ Pulled %s → %s\n", args[0], outPath)
		return nil
	},
}

func init() {
	registryCmd.AddCommand(registryPullCmd)
}
