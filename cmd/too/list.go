package main

import (
	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/commands/datapath"
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
		rawCollectionPath, _ := cmd.Flags().GetString("data-path")
		collectionPath := datapath.ResolveCollectionPath(rawCollectionPath)

		// Call business logic using unified command
		opts := map[string]interface{}{
			"collectionPath": collectionPath,
			"done":           showDone,
			"all":            showAll,
		}
		result, err := too.ExecuteUnifiedCommand("list", []string{}, opts)
		if err != nil {
			return err
		}

		// Render output
		renderer, err := getRenderer()
		if err != nil {
			return err
		}
		
		return renderer.RenderChange(result)
	},
}

func init() {
	// Add flags for filtering
	listCmd.Flags().BoolVarP(&showDone, "done", "d", false, msgFlagDone)
	listCmd.Flags().BoolVarP(&showAll, "all", "a", false, msgFlagAll)

	rootCmd.AddCommand(listCmd)
}
