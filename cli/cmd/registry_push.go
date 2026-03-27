package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/aallbrig/beetree-cli/internal/registry"
	"github.com/spf13/cobra"
)

var (
	pushPublic  bool
	pushPrivate bool
	pushTags    []string
	pushDesc    string
)

var registryPushCmd = &cobra.Command{
	Use:   "push <file>",
	Short: "Publish a behavior tree spec to the registry",
	Long:  "Upload a .beetree.yaml spec to the BeeTree registry for sharing.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		specFile := args[0]
		data, err := os.ReadFile(specFile)
		if err != nil {
			return fmt.Errorf("read spec: %w", err)
		}

		public := true
		if pushPrivate {
			public = false
		}

		client := getRegistryClient()
		entry, err := client.Push(context.Background(), data, registry.PushOptions{
			Public:      public,
			Description: pushDesc,
			Tags:        pushTags,
		})
		if err != nil {
			return fmt.Errorf("push failed: %w", err)
		}

		visibility := "public"
		if !entry.Public {
			visibility = "private"
		}
		cmd.Printf("✓ Published %s (%s)\n", entry.FullName(), visibility)
		return nil
	},
}

func init() {
	registryPushCmd.Flags().BoolVar(&pushPublic, "public", true, "Make tree publicly visible")
	registryPushCmd.Flags().BoolVar(&pushPrivate, "private", false, "Make tree private")
	registryPushCmd.Flags().StringSliceVar(&pushTags, "tag", nil, "Tags for the tree (can be specified multiple times)")
	registryPushCmd.Flags().StringVar(&pushDesc, "description", "", "Description of the tree")
	registryCmd.AddCommand(registryPushCmd)
}
