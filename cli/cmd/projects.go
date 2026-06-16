package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/0DayMonxrch/vaultify/cli/client"
	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewClient()
		if err != nil {
			return err
		}

		projects, err := c.ListProjects()
		if err != nil {
			return err
		}

		if len(projects) == 0 {
			cmd.Println("No projects found.")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSLUG\tCREATED AT")
		for _, p := range projects {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, p.Name, p.Slug, p.CreatedAt)
		}
		w.Flush()
		return nil
	},
}

func init() {
	projectsCmd.AddCommand(projectsListCmd)
	RootCmd.AddCommand(projectsCmd)
}
