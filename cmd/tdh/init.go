package main

import (
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     msgInitUse,
	Aliases: aliasesInit,
	Short:   msgInitShort,
	Long:    msgInitLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic
		result, err := tdh.Init(tdh.InitOptions{
			DBPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer := output.NewRenderer(nil)
		return renderer.RenderInit(result)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
