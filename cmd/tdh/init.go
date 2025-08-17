package main

import (
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/spf13/cobra"
)

var initUseHomeDir bool

var initCmd = &cobra.Command{
	Use:     msgInitUse,
	Aliases: aliasesInit,
	Short:   msgInitShort,
	Long:    msgInitLong,
	GroupID: "misc",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic
		result, err := tdh.Init(tdh.InitOptions{
			DBPath:     collectionPath,
			UseHomeDir: initUseHomeDir,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer, err := getRenderer()
		if err != nil {
			return err
		}
		return renderer.RenderInit(result)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initUseHomeDir, "home", false, "Create .todos file in home directory instead of current directory")
}
