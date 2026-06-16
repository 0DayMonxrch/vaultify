package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"github.com/0DayMonxrch/vaultify/cli/client"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run --project <slug> --env <env> -- <command>",
	Short: "Run a subprocess with decrypted secrets injected into its environment",
	Args:  cobra.MinimumNArgs(1),
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

		secretsMeta, err := c.ListSecrets(targetProjectID, secretEnv)
		if err != nil {
			return err
		}

		// Fetch decrypted secrets
		var envVars []string
		for _, sm := range secretsMeta {
			val, err := c.GetDecryptedSecret(targetProjectID, sm.ID)
			if err != nil {
				return fmt.Errorf("failed to fetch secret %s: %w", sm.KeyName, err)
			}
			envVars = append(envVars, fmt.Sprintf("%s=%s", sm.KeyName, val))
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle interrupts
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		go func() {
			<-sigChan
			cancel()
		}()

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
		envVars = nil

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
	runCmd.MarkFlagRequired("project")

	RootCmd.AddCommand(runCmd)
}
