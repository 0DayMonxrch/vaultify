package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:           "vaultify",
	Short:         "Vaultify CLI injects secrets into your environment.",
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		green := color.New(color.FgGreen, color.Bold).SprintFunc()

		fmt.Println(cyan(`
 __      __         _  _    _  __        
 \ \    / /        | || |  (_)/ _|       
  \ \  / /_ _ _   _| || |_  _| |_ _   _  
   \ \/ / _' | | | | || __|| |  _| | | | 
    \  / (_| | |_| | || |_ | | | | |_| | 
     \/ \__,_|\__,_|_| \__||_|_|  \__, | 
                                   __/ | 
                                  |___/  `))
		fmt.Printf("%s\n", cyan("Vaultify - Secure Secrets Injection CLI"))
		fmt.Printf("Seamlessly inject decrypted secrets directly into your application's environment.\n\n")

		fmt.Printf("Developed by: %s\n", green("Dibyadipan"))
		fmt.Printf("GitHub: %s\n\n", yellow("https://github.com/0DayMonxrch"))

		fmt.Printf("%s\n", cyan("Quick Start Guide:"))
		fmt.Printf("  1. Authenticate with your token: %s\n", green("vaultify login --token <vt_...>"))
		fmt.Printf("  2. List available projects:      %s\n", green("vaultify projects list"))
		fmt.Printf("  3. Run your application:         %s\n\n", green("vaultify run --project <slug> -- <command>"))

		fmt.Printf("Run %s for more information about commands.\n", green("vaultify --help"))
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		color.Red("✖ Error: %v\n", err)
		os.Exit(1)
	}
}
