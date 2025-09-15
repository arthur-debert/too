package main

import (
	"github.com/arthur-debert/too/pkg/too/commands/datapath"
	"github.com/spf13/cobra"

	"github.com/arthur-debert/too/pkg/too"
)

var moveCmd = &cobra.Command{
	Use:     msgMoveUse,
	Aliases: aliasesMove,
	Short:   msgMoveShort,
	Long:    msgMoveLong,
	GroupID: "extras",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		destParentPath := args[1]

		// Get collection path from command flags
		rawCollectionPath, _ := cmd.Flags().GetString("data-path")
		collectionPath := datapath.ResolveCollectionPath(rawCollectionPath)

		// Call business logic using unified command
		opts := map[string]interface{}{
			"collectionPath": collectionPath,
		}
		result, err := too.ExecuteUnifiedCommand("move", []string{sourcePath, destParentPath}, opts)
		if err != nil {
			return err
		}

		// Render output
		return renderToStdout(result)
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)
}
