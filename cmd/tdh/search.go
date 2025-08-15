package main

import (
	"strings"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
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
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Get case-sensitive flag
		caseSensitive, _ := cmd.Flags().GetBool("case-sensitive")

		// Call business logic
		result, err := tdh.Search(query, tdh.SearchOptions{
			CollectionPath: collectionPath,
			CaseSensitive:  caseSensitive,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer := output.NewRenderer(nil)
		return renderer.RenderSearch(result)
	},
}

func init() {
	searchCmd.Flags().BoolP("case-sensitive", "s", false, msgFlagCaseSensitive)
	rootCmd.AddCommand(searchCmd)
}
