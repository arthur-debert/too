package main

import (
	"github.com/spf13/cobra"

	"github.com/arthur-debert/too/pkg/too"
)

var moveCmd = &cobra.Command{
	Use:     msgMoveUse,
	Aliases: aliasesMove,
	Short:   msgMoveShort,
	Long:    msgMoveLong,
	GroupID: "extras",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		destParentPath := args[1]

		// Get collection path from command flags
		rawCollectionPath, _ := cmd.Flags().GetString("data-path")
		collectionPath := too.ResolveCollectionPath(rawCollectionPath)

		result, err := too.Move(sourcePath, destParentPath, too.MoveOptions{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		renderer, err := getRenderer()
		if err != nil {
			return err
		}
		return renderer.RenderMove(result)
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)
}
