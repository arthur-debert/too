package main

import (
	"strconv"
	"strings"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/display"
	"github.com/spf13/cobra"
)

var modifyCmd = &cobra.Command{
	Use:     "modify <id> <text>",
	Aliases: []string{"m"},
	Short:   "Modify the text of an existing todo",
	Long:    `Modify the text of an existing todo by its ID.`,
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
		collectionPath, _ := cmd.Flags().GetString("collection")

		// Call business logic
		result, err := tdh.Modify(id, text, tdh.ModifyOptions{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer := display.NewRenderer(nil)
		return renderer.RenderModify(result)
	},
}

func init() {
	rootCmd.AddCommand(modifyCmd)
}
