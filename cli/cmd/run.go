package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/0DayMonxrch/vaultify/cli/client"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run --project <slug> --env <env> -- <command>",
	Short: "Run a subprocess with decrypted secrets injected into its environment",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Fetching and decrypting secrets..."
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

		secretsMeta, err := c.ListSecrets(targetProjectID, secretEnv)
		if err != nil {
			s.Stop()
			return err
		}

		// Fetch decrypted secrets
		var envVars []string
		for _, sm := range secretsMeta {
			val, err := c.GetDecryptedSecret(targetProjectID, sm.ID)
			if err != nil {
				s.Stop()
				return fmt.Errorf("failed to fetch secret %s: %w", sm.KeyName, err)
			}
			envVars = append(envVars, fmt.Sprintf("%s=%s", sm.KeyName, val))
		}

		s.Stop()
		color.Green("✓ Secrets injected successfully. Starting process...\n")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle interrupts
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		go func() {
			<-sigChan
			cancel()
		}()

		// #nosec G204
		proc := exec.CommandContext(ctx, args[0], args[1:]...)
		proc.Env = append(os.Environ(), envVars...)
		proc.Stdout = cmd.OutOrStdout()
		proc.Stderr = cmd.ErrOrStderr()
		proc.Stdin = cmd.InOrStdin()

		err = proc.Start()
		if err != nil {
			return fmt.Errorf("failed to start command: %w", err)
		}

		// Best-effort memory zeroing. Relies on Go GC to eventually clean up
		// the memory since strings are immutable in Go, but we clear the references.
		for i := range envVars {
			envVars[i] = ""
		}

		err = proc.Wait()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				os.Exit(exitError.ExitCode())
			}
			return fmt.Errorf("command execution failed: %w", err)
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringVar(&projectSlug, "project", "", "Project slug")
	runCmd.Flags().StringVar(&secretEnv, "env", "", "Environment name")
	_ = runCmd.MarkFlagRequired("project")

	RootCmd.AddCommand(runCmd)
}
