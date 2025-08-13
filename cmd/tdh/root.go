package main

import (
	"fmt"

	"github.com/arthur-debert/tdh/internal/version"
	"github.com/arthur-debert/tdh/pkg/logging"
	"github.com/arthur-debert/tdh/pkg/models"
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/display"

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
		// Default action is to list todos
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get collection path from flag
			collectionPath, _ := cmd.Flags().GetString("collection")
			showDone, _ := cmd.Flags().GetBool("done")
			showAll, _ := cmd.Flags().GetBool("all")

			// Call business logic
			result, err := tdh.List(tdh.ListOptions{
				CollectionPath: collectionPath,
			})
			if err != nil {
				return err
			}

			// Filter based on flags
			if !showAll {
				var filteredTodos []*models.Todo
				for _, todo := range result.Todos {
					if showDone && todo.Status == "done" {
						filteredTodos = append(filteredTodos, todo)
					} else if !showDone && todo.Status != "done" {
						filteredTodos = append(filteredTodos, todo)
					}
				}
				result.Todos = filteredTodos
			}

			// Render output
			renderer := display.NewRenderer(nil)
			return renderer.RenderList(result)
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
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

	// Add global flags for filtering (these work with the default list command)
	rootCmd.Flags().BoolP("done", "d", false, "print done todos")
	rootCmd.Flags().BoolP("all", "a", false, "print all todos")

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
