package main

import (
	cmdInit "github.com/arthur-debert/too/pkg/too/commands/init"
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
		result, err := cmdInit.Execute(cmdInit.Options{
			DBPath:     collectionPath,
			UseHomeDir: initUseHomeDir,
		})
		if err != nil {
			return err
		}

		// Render output
		return renderToStdout(result)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initUseHomeDir, "home", false, "Create .todos file in home directory instead of current directory")
}
