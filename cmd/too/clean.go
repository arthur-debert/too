package main

import (
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

		// Call business logic using unified command
		opts := map[string]interface{}{
			"collectionPath": collectionPath,
		}
		result, err := too.ExecuteUnifiedCommand("clean", []string{}, opts)
		if err != nil {
			return err
		}

		// Render output
		renderer, err := getRenderer()
		if err != nil {
			return err
		}
		
		return renderer.RenderChange(result)
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
