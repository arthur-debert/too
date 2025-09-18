package main

import (
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
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If root command is called directly, run list
			return listCmd.RunE(listCmd, args)
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Setup logging based on verbosity
			logging.SetupLogger(verbosity)
			log.Debug().Str("command", cmd.Name()).Msg("Command started")
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	// Try to execute normally first
	err := rootCmd.Execute()
	
	// If we got "unknown command" error, it might be naked execution
	if err != nil && isUnknownCommandError(err) {
		// Handle naked execution by injecting appropriate command
		if handleErr := handleNakedExecution(); handleErr != nil {
			return err // Return original error if we can't handle it
		}
		
		// Try executing again with the injected command
		return rootCmd.Execute()
	}
	
	return err
}

func init() {
	// Set version template
	rootCmd.SetVersionTemplate(msgRootVersion)

	// Define command groups
	rootCmd.AddGroup(
		&cobra.Group{ID: "core", Title: "CORE:"},
		&cobra.Group{ID: "extras", Title: "EXTRAS:"},
		&cobra.Group{ID: "misc", Title: "MISC:"},
	)

	// Persistent flags
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", msgFlagVerbose)
	rootCmd.PersistentFlags().StringP("data-path", "p", "", msgFlagDataPath)
	rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "term", msgFlagFormat)

	// Setup custom help
	setupHelp()
}