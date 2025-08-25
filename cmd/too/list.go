package main

import (
	"github.com/arthur-debert/too/pkg/too"
	"github.com/spf13/cobra"
)

var (
	showDone bool
	showAll  bool
)

var listCmd = &cobra.Command{
	Use:     msgListUse,
	Aliases: aliasesList,
	Short:   msgListShort,
	Long:    msgListLong,
	GroupID: "extras",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic with filtering options
		result, err := too.List(too.ListOptions{
			CollectionPath: collectionPath,
			ShowDone:       showDone,
			ShowAll:        showAll,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer, err := getRenderer()
		if err != nil {
			return err
		}
		return renderer.RenderList(result)
	},
}

func init() {
	// Add flags for filtering
	listCmd.Flags().BoolVarP(&showDone, "done", "d", false, msgFlagDone)
	listCmd.Flags().BoolVarP(&showAll, "all", "a", false, msgFlagAll)

	rootCmd.AddCommand(listCmd)
}
