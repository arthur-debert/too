package main

import (
	"strconv"
	"strings"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:     "edit <id> <text>",
	Aliases: []string{"modify", "m", "e"},
	Short:   "Edit the text of an existing todo",
	Long:    `Edit the text of an existing todo by its ID.`,
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse ID
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		// Join remaining arguments as the new text
		text := strings.Join(args[1:], " ")

		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic
		result, err := tdh.Modify(id, text, tdh.ModifyOptions{
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
