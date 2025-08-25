package main

import (
	"github.com/arthur-debert/too/pkg/too"
	"github.com/spf13/cobra"
)

var datapathCmd = &cobra.Command{
	Use:     "datapath",
	Aliases: []string{"path"},
	Short:   "Show the path to the todo data file",
	Long:    "Display the full path to the todo data file currently being used.",
	GroupID: "misc",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic
		result, err := too.ShowDataPath(too.ShowDataPathOptions{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer, err := getRenderer()
		if err != nil {
			return err
		}
		return renderer.RenderDataPath(result)
	},
}

func init() {
	rootCmd.AddCommand(datapathCmd)
}
