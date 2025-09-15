package main

import (
	"github.com/arthur-debert/too/pkg/too/commands/datapath"
	"github.com/arthur-debert/too/pkg/too"
	"github.com/spf13/cobra"
)

var reopenCmd = &cobra.Command{
	Use:     msgReopenUse,
	Aliases: aliasesReopen,
	Short:   msgReopenShort,
	Long:    msgReopenLong,
	GroupID: "core",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		rawCollectionPath, _ := cmd.Flags().GetString("data-path")
		collectionPath := datapath.ResolveCollectionPath(rawCollectionPath)

		// Call business logic using unified command
		opts := map[string]interface{}{
			"collectionPath": collectionPath,
		}
		result, err := too.ExecuteUnifiedCommand("reopen", args, opts)
		if err != nil {
			return err
		}

		// Render output
		return renderToStdout(result)
	},
}

func init() {
	rootCmd.AddCommand(reopenCmd)
}
