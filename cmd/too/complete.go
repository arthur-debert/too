package main

import (
	"github.com/arthur-debert/too/pkg/too"
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
		var results []*too.CompleteResult
		for _, positionPath := range args {
			result, err := too.Complete(positionPath, too.CompleteOptions{
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
