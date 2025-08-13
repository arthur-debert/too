package main

import (
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/display"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/spf13/cobra"
)

var (
	showDone bool
	showAll  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all todos",
	Long:  `List all todos in the collection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("collection")

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

func init() {
	// Add flags for filtering
	listCmd.Flags().BoolVarP(&showDone, "done", "d", false, "print done todos")
	listCmd.Flags().BoolVarP(&showAll, "all", "a", false, "print all todos")

	rootCmd.AddCommand(listCmd)
}
