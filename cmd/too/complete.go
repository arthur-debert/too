package main

import (
	"fmt"
	
	"github.com/arthur-debert/too/pkg/too/commands/datapath"
	"github.com/arthur-debert/too/pkg/too"
	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:     msgCompleteUse,
	Aliases: aliasesComplete,
	Short:   msgCompleteShort,
	Long:    msgCompleteLong,
	GroupID: "core",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		collectionPath := resolveDataPath(cmd)
		
		// Ensure gitignore is updated for project scope
		if err := datapath.EnsureProjectGitignore(); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: could not update .gitignore: %v\n", err)
		}

		// Call business logic using unified command
		opts := map[string]interface{}{
			"collectionPath": collectionPath,
		}
		result, err := too.ExecuteUnifiedCommand("complete", args, opts)
		if err != nil {
			return err
		}

		// Render output
		return renderToStdout(result)
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
}
