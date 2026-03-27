package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/registry"
	"github.com/spf13/cobra"
)

var (
	browseTag  string
	browseSort string
)

var registryBrowseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse public behavior trees in the registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getRegistryClient()
		entries, err := client.Browse(context.Background(), registry.BrowseOptions{
			Tag:   browseTag,
			Sort:  browseSort,
			Limit: 50,
		})
		if err != nil {
			return fmt.Errorf("browse failed: %w", err)
		}
		if len(entries) == 0 {
			cmd.Println("No trees found.")
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
	registryBrowseCmd.Flags().StringVar(&browseTag, "tag", "", "Filter by tag")
	registryBrowseCmd.Flags().StringVar(&browseSort, "sort", "recent", "Sort order: popular, recent, name")
	registryCmd.AddCommand(registryBrowseCmd)
}
