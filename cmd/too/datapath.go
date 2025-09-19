package main

import (
	"github.com/arthur-debert/too/pkg/too/commands/datapath"
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
		collectionPath := resolveDataPath(cmd)

		// Call business logic using datapath package
		result, err := datapath.Execute(datapath.Options{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		return renderToStdout(result)
	},
}

func init() {
	rootCmd.AddCommand(datapathCmd)
}
