package cmd

import (
	"fmt"

	"github.com/0DayMonxrch/vaultify/cli/config"
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
		if token == "" || host == "" {
			return fmt.Errorf("--token and --host are required")
		}
		cfg := &config.Config{
			Host:  host,
			Token: token,
		}
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		cmd.Println("Successfully logged in.")
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&host, "host", "", "Vaultify server URL")
	loginCmd.Flags().StringVar(&token, "token", "", "API Token (vt_...)")
	RootCmd.AddCommand(loginCmd)
}
