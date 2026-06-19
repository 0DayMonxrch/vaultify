package cmd

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/0DayMonxrch/vaultify/cli/client"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
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
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Fetching secrets metadata..."
		s.Start()

		c, err := client.NewClient()
		if err != nil {
			s.Stop()
			return err
		}

		projects, err := c.ListProjects()
		if err != nil {
			s.Stop()
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
			s.Stop()
			return fmt.Errorf("project with slug %q not found", projectSlug)
		}

		secrets, err := c.ListSecrets(targetProjectID, secretEnv)
		s.Stop()

		if err != nil {
			return err
		}

		if len(secrets) == 0 {
			color.Yellow("No secrets found.")
			return nil
		}

		color.Cyan("\nPROJECT SECRETS")
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 4, ' ', 0)
		_, _ = fmt.Fprintln(w, "KEY\tENVIRONMENT\tCREATED AT\tUPDATED AT")
		for _, s := range secrets {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.KeyName, s.Environment, s.CreatedAt, s.UpdatedAt)
		}
		_ = w.Flush()
		return nil
	},
}

func init() {
	secretsListCmd.Flags().StringVar(&projectSlug, "project", "", "Project slug")
	secretsListCmd.Flags().StringVar(&secretEnv, "env", "", "Environment name")
	_ = secretsListCmd.MarkFlagRequired("project")

	secretsCmd.AddCommand(secretsListCmd)
	RootCmd.AddCommand(secretsCmd)
}
