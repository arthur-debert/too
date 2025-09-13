package main

import (
	"os"

	"github.com/arthur-debert/too/internal/version"
	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too"

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
func Execute() error {
	// Check if a subcommand was provided
	args := os.Args[1:]
	hasCommand := false
	hasHelpFlag := false
	hasVersionFlag := false

	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			hasHelpFlag = true
			break
		}
		if arg == "--version" {
			hasVersionFlag = true
			break
		}
		if len(arg) > 0 && arg[0] != '-' {
			hasCommand = true
			break
		}
	}

	// Default to list if no command and no help/version flag
	if !hasCommand && !hasHelpFlag && !hasVersionFlag {
		os.Args = append([]string{os.Args[0], "list"}, os.Args[1:]...)
	}

	return rootCmd.Execute()
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

	// Register unified add command as a test
	if addDef, ok := too.CommandRegistry["add"]; ok {
		addCmd := createUnifiedCommand(addDef)
		addCmd.GroupID = "core"
		addCmd.Use = "add-unified <text>"
		addCmd.Aliases = []string{"au"}
		rootCmd.AddCommand(addCmd)
	}
}