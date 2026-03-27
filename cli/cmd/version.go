package cmd

import (
	"github.com/spf13/cobra"
)

var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of beetree",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("beetree version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
