package main

import (
	"fmt"
	"os"

	"github.com/arthur-debert/tdh/internal/version"
	"github.com/arthur-debert/tdh/pkg/logging"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	verbosity int

	rootCmd = &cobra.Command{
		Use:   "tdh",
		Short: "A simple command-line todo list manager",
		Long: `tdh is a simple command-line todo list manager that helps you track tasks.
It stores todos in a JSON file and provides commands to add, modify, toggle, and search todos.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Setup logging based on verbosity
			logging.SetupLogger(verbosity)
			log.Debug().Str("command", cmd.Name()).Msg("Command started")
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	// Check if a subcommand was provided
	// If not, default to list command (unless help/version is requested)
	args := os.Args[1:] // Skip program name
	hasCommand := false
	hasHelpFlag := false

	for _, arg := range args {
		// Check for help flags
		if arg == "-h" || arg == "--help" || arg == "help" {
			hasHelpFlag = true
			break
		}
		// If it's not a flag (doesn't start with -), it might be a command
		if len(arg) > 0 && arg[0] != '-' {
			hasCommand = true
			break
		}
	}

	// Only default to list if no command and no help flag
	if !hasCommand && !hasHelpFlag {
		// Insert "list" after program name but before any flags
		os.Args = append([]string{os.Args[0], "list"}, os.Args[1:]...)
	}

	return rootCmd.Execute()
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Verbosity flag for logging
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase verbosity (-v, -vv, -vvv)")

	// Add persistent flags
	rootCmd.PersistentFlags().StringP("collection", "c", "", "path to todo collection (default: $HOME/.todos.json)")

	// Add version command
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of tdh`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("tdh version %s\n", version.Info())

	},
}
