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

		// Render output
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
