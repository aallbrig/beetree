package cmd

import (
	"os"
	"path/filepath"

	"github.com/aallbrig/beetree-cli/internal/registry"
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:    "registry",
	Short:  "Interact with the BeeTree registry",
	Long:   "Browse, search, pull, and push behavior tree specs to the BeeTree registry.",
	Hidden: true, // Registry backend is not yet implemented; hide from help until ready
}

// registryClientOverride allows tests to inject a custom client.
var registryClientOverride registry.Client

func getRegistryDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".beetree", "registry")
}

func getRegistryClient() registry.Client {
	if registryClientOverride != nil {
		return registryClientOverride
	}
	return registry.NewLocalClient(getRegistryDir())
}

func init() {
	rootCmd.AddCommand(registryCmd)
}
