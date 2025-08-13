package main

import (
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/display"
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

		// Call business logic with filtering options
		result, err := tdh.List(tdh.ListOptions{
			CollectionPath: collectionPath,
			ShowDone:       showDone,
			ShowAll:        showAll,
		})
		if err != nil {
			return err
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
