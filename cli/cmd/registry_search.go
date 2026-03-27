package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/registry"
	"github.com/spf13/cobra"
)

var registrySearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for behavior trees in the registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getRegistryClient()
		entries, err := client.Search(context.Background(), registry.SearchOptions{
			Query: args[0],
			Limit: 50,
		})
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}
		if len(entries) == 0 {
			cmd.Println("No trees found matching query.")
			return nil
		}
		for _, e := range entries {
			tags := ""
			if len(e.Tags) > 0 {
				tags = " [" + strings.Join(e.Tags, ", ") + "]"
			}
			cmd.Printf("  %s — %s%s\n", e.FullName(), e.Description, tags)
		}
		cmd.Printf("\n%d tree(s) found.\n", len(entries))
		return nil
	},
}

func init() {
	registryCmd.AddCommand(registrySearchCmd)
}
