package main

import (
	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
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
		collectionPath := too.ResolveCollectionPath(rawCollectionPath)

		// Call business logic for each position path
		var results []*too.ReopenResult
		for _, positionPath := range args {
			result, err := too.Reopen(positionPath, too.ReopenOptions{
				CollectionPath: collectionPath,
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
		
		if len(results) == 0 {
			return nil
		}
		
		// Collect all affected todos
		affectedTodos := make([]*models.Todo, len(results))
		for i, result := range results {
			affectedTodos[i] = result.Todo
		}
		
		// Use data from the last result
		lastResult := results[len(results)-1]
		changeResult := too.NewChangeResult(
			"placeholder",
			"reopened",
			affectedTodos,
			lastResult.AllTodos,
			lastResult.TotalCount,
			lastResult.DoneCount,
		)
		
		return renderer.RenderChange(changeResult)
	},
}

func init() {
	rootCmd.AddCommand(reopenCmd)
}
