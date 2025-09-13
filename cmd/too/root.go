package main

import (
	"os"

	"github.com/arthur-debert/too/internal/version"
	"github.com/arthur-debert/too/pkg/logging"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	verbosity  int
	formatFlag string

	rootCmd = &cobra.Command{
		Use:     "too",
		Short:   msgRootShort,
		Long:    msgRootLong,
		Version: version.Info(),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
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
	hasVersionFlag := false

	for _, arg := range args {
		// Check for help flags
		if arg == "-h" || arg == "--help" || arg == "help" {
			hasHelpFlag = true
			break
		}
		// Check for version flag
		if arg == "--version" {
			hasVersionFlag = true
			break
		}
		// If it's not a flag (doesn't start with -), it might be a command
		if len(arg) > 0 && arg[0] != '-' {
			hasCommand = true
			break
		}
	}

	// Only default to list if no command and no help/version flag
	if !hasCommand && !hasHelpFlag && !hasVersionFlag {
		// Insert "list" after program name but before any flags
		os.Args = append([]string{os.Args[0], "list"}, os.Args[1:]...)
	}

	return rootCmd.Execute()
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Set version template
	rootCmd.SetVersionTemplate(msgRootVersion)

	// Define command groups
	rootCmd.AddGroup(
		&cobra.Group{ID: "core", Title: "CORE:"},
		&cobra.Group{ID: "extras", Title: "EXTRAS:"},
		&cobra.Group{ID: "misc", Title: "MISC:"},
	)

	// Verbosity flag for logging
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", msgFlagVerbose)

	// Add persistent flags
	rootCmd.PersistentFlags().StringP("data-path", "p", "", msgFlagDataPath)
	rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "term", msgFlagFormat)

	// Setup custom help
	setupHelp()
}
