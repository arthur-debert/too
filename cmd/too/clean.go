package main

import (
	"fmt"
	
	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/commands/datapath"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:     msgCleanUse,
	Short:   msgCleanShort,
	Long:    msgCleanLong,
	GroupID: "misc",
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
		result, err := too.ExecuteUnifiedCommand("clean", []string{}, opts)
		if err != nil {
			return err
		}

		// Render output
		return renderToStdout(result)
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
