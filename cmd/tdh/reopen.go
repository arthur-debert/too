package main

import (
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
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
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic for each position path
		var results []*tdh.ReopenResult
		for _, positionPath := range args {
			result, err := tdh.Reopen(positionPath, tdh.ReopenOptions{
				CollectionPath: collectionPath,
			})
			if err != nil {
				return err
			}
			results = append(results, result)
		}

		// Render output
		renderer := output.NewRenderer(nil)
		return renderer.RenderReopen(results)
	},
}

func init() {
	rootCmd.AddCommand(reopenCmd)
}
