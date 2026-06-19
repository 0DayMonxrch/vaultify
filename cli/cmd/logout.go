package cmd

import (
	"fmt"

	"github.com/0DayMonxrch/vaultify/cli/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the locally stored configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.DeleteConfig(); err != nil {
			return fmt.Errorf("failed to logout: %w", err)
		}
		color.Green("✓ Successfully logged out.")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}
