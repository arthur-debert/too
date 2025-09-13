package main

import (
	"strings"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     msgSearchUse,
	Aliases: aliasesSearch,
	Short:   msgSearchShort,
	Long:    msgSearchLong,
	GroupID: "extras",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Join all arguments as the search query
		query := strings.Join(args, " ")

		// Get collection path from flag
		rawCollectionPath, _ := cmd.Flags().GetString("data-path")
		collectionPath := too.ResolveCollectionPath(rawCollectionPath)

		// Get case-sensitive flag
		caseSensitive, _ := cmd.Flags().GetBool("case-sensitive")

		// Call business logic
		result, err := too.Search(query, too.SearchOptions{
			CollectionPath: collectionPath,
			CaseSensitive:  caseSensitive,
		})
		if err != nil {
			return err
		}

		// Get unified change result
		unifiedResult, err := too.ExecuteUnifiedCommand("search", []string{query}, map[string]interface{}{
			"collectionPath": collectionPath,
			"query":          query,
		})
		if err == nil {
			// Use unified renderer
			renderer, err := getRenderer()
			if err != nil {
				return err
			}
			return renderer.RenderChange(unifiedResult)
		}
		
		// Fallback to legacy renderer
		renderer, err := getRenderer()
		if err != nil {
			return err
		}
		return renderer.RenderSearch(result)
	},
}

func init() {
	searchCmd.Flags().BoolP("case-sensitive", "s", false, msgFlagCaseSensitive)
	rootCmd.AddCommand(searchCmd)
}
