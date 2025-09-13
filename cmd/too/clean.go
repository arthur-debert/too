package main

import (
	"fmt"
	
	"github.com/arthur-debert/too/pkg/too"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:     msgCleanUse,
	Short:   msgCleanShort,
	Long:    msgCleanLong,
	GroupID: "misc",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		rawCollectionPath, _ := cmd.Flags().GetString("data-path")
		collectionPath := too.ResolveCollectionPath(rawCollectionPath)

		// Call business logic
		result, err := too.Clean(too.CleanOptions{
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
		
		// Get unified change result
		unifiedResult, err := too.ExecuteUnifiedCommand("clean", []string{}, map[string]interface{}{
			"collectionPath": collectionPath,
		})
		if err != nil {
			// Fallback to legacy result if unified command fails
			changeResult := too.NewChangeResult(
				"clean",
				fmt.Sprintf("Cleaned %d todos", len(result.RemovedTodos)),
				result.RemovedTodos,
				result.ActiveTodos,
				result.ActiveCount,
				0, // After clean, no done todos remain
			)
			return renderer.RenderChange(changeResult)
		}
		
		return renderer.RenderChange(unifiedResult)
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
