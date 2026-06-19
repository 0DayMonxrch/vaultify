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

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Fetching projects..."
		s.Start()

		c, err := client.NewClient()
		if err != nil {
			s.Stop()
			return err
		}

		projects, err := c.ListProjects()
		s.Stop()

		if err != nil {
			return err
		}

		if len(projects) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), color.YellowString("No projects found."))
			return nil
		}

		fmt.Fprintln(cmd.OutOrStdout(), color.CyanString("\nAVAILABLE PROJECTS"))
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 4, ' ', 0)
		_, _ = fmt.Fprintln(w, "ID\tNAME\tSLUG\tCREATED AT")
		for _, p := range projects {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, p.Name, p.Slug, p.CreatedAt)
		}
		_ = w.Flush()
		return nil
	},
}

func init() {
	projectsCmd.AddCommand(projectsListCmd)
	RootCmd.AddCommand(projectsCmd)
}
