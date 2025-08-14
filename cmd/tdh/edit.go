package main

import (
	"strconv"
	"strings"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:     msgEditUse,
	Aliases: aliasesEdit,
	Short:   msgEditShort,
	Long:    msgEditLong,
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse position
		position, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		// Join remaining arguments as the new text
		text := strings.Join(args[1:], " ")

		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic
		result, err := tdh.Modify(position, text, tdh.ModifyOptions{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer := output.NewRenderer(nil)
		return renderer.RenderModify(result)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
