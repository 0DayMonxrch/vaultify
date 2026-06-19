package cmd

import (
	"fmt"

	"github.com/0DayMonxrch/vaultify/cli/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	host  string
	token string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save the API token and host configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if token == "" {
			return fmt.Errorf("--token is required")
		}

		targetHost := host
		if targetHost == "" {
			targetHost = "https://try-vaultify.tech"
		}

		cfg := &config.Config{
			Host:  targetHost,
			Token: token,
		}
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("✓ Successfully logged in."))
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&host, "host", "", "Vaultify server URL (default: https://try-vaultify.tech)")
	loginCmd.Flags().StringVar(&token, "token", "", "API Token (vt_...)")
	RootCmd.AddCommand(loginCmd)
}
