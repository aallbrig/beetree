package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var registryLoginCmd = &cobra.Command{
	Use:   "login <token>",
	Short: "Authenticate with the BeeTree registry",
	Long:  "Store a GitHub personal access token for registry operations.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getRegistryClient()
		if err := client.Login(context.Background(), args[0]); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
		cmd.Println("✓ Logged in to BeeTree registry")
		return nil
	},
}

var registryLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored registry credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getRegistryClient()
		if err := client.Logout(context.Background()); err != nil {
			return fmt.Errorf("logout failed: %w", err)
		}
		cmd.Println("✓ Logged out of BeeTree registry")
		return nil
	},
}

func init() {
	registryCmd.AddCommand(registryLoginCmd)
	registryCmd.AddCommand(registryLogoutCmd)
}
