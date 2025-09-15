package main

import (
	"strings"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/commands/datapath"
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
		collectionPath := datapath.ResolveCollectionPath(rawCollectionPath)

		// Get case-sensitive flag (not currently used in unified command)
		// caseSensitive, _ := cmd.Flags().GetBool("case-sensitive")

		// Call business logic using unified command
		opts := map[string]interface{}{
			"collectionPath": collectionPath,
			"query":          query,
		}
		result, err := too.ExecuteUnifiedCommand("search", []string{query}, opts)
		if err != nil {
			return err
		}

		// Render output
		return renderToStdout(result)
	},
}

func init() {
	searchCmd.Flags().BoolP("case-sensitive", "s", false, msgFlagCaseSensitive)
	rootCmd.AddCommand(searchCmd)
}
