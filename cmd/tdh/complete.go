package main

import (
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:     msgCompleteUse,
	Aliases: aliasesComplete,
	Short:   msgCompleteShort,
	Long:    msgCompleteLong,
	GroupID: "core",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic for each position path
		var results []*tdh.CompleteResult
		for _, positionPath := range args {
			result, err := tdh.Complete(positionPath, tdh.CompleteOptions{
				CollectionPath: collectionPath,
				Mode:           modeFlag,
			})
			if err != nil {
				return err
			}
			results = append(results, result)
		}

		// Render output
		renderer, err := getRenderer()
		if err != nil {
			return err
		}
		return renderer.RenderComplete(results)
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
}
