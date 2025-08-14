package main

import (
	"strings"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/spf13/cobra"
)

var (
	parentPath string
)

var addCmd = &cobra.Command{
	Use:     msgAddUse,
	Aliases: aliasesAdd,
	Short:   msgAddShort,
	Long:    msgAddLong,
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Join all arguments as the todo text
		text := strings.Join(args, " ")

		// Get flags
		collectionPath, _ := cmd.Flags().GetString("data-path")
		parentPath, _ := cmd.Flags().GetString("parent")

		// Call business logic
		result, err := tdh.Add(text, tdh.AddOptions{
			CollectionPath: collectionPath,
			ParentPath:     parentPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer := output.NewRenderer(nil)
		return renderer.RenderAdd(result)
	},
}

func init() {
	addCmd.Flags().StringVar(&parentPath, "parent", "", "parent todo position path (e.g., \"1.2\")")
	rootCmd.AddCommand(addCmd)
}
