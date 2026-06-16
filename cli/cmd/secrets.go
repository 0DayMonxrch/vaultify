package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/0DayMonxrch/vaultify/cli/client"
	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage secrets",
}

var (
	projectSlug string
	secretEnv   string
)

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secret keys in a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient()
		if err != nil {
			return err
		}

		projects, err := c.ListProjects()
		if err != nil {
			return err
		}

		var targetProjectID string
		for _, p := range projects {
			if p.Slug == projectSlug {
				targetProjectID = p.ID
				break
			}
		}

		if targetProjectID == "" {
			return fmt.Errorf("project with slug %q not found", projectSlug)
		}

		secrets, err := c.ListSecrets(targetProjectID, secretEnv)
		if err != nil {
			return err
		}

		if len(secrets) == 0 {
			cmd.Println("No secrets found.")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "KEY\tENVIRONMENT\tCREATED AT\tUPDATED AT")
		for _, s := range secrets {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.KeyName, s.Environment, s.CreatedAt, s.UpdatedAt)
		}
		w.Flush()
		return nil
	},
}

func init() {
	secretsListCmd.Flags().StringVar(&projectSlug, "project", "", "Project slug")
	secretsListCmd.Flags().StringVar(&secretEnv, "env", "", "Environment name")
	secretsListCmd.MarkFlagRequired("project")

	secretsCmd.AddCommand(secretsListCmd)
	RootCmd.AddCommand(secretsCmd)
}
