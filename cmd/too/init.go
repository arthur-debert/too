package main

import (
	"fmt"
	cmdInit "github.com/arthur-debert/too/pkg/too/commands/init"
	"github.com/arthur-debert/too/pkg/too/commands/datapath"
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
		explicitPath, _ := cmd.Flags().GetString("data-path")
		isGlobal, _ := cmd.Flags().GetBool("global")
		
		// For init, we need special handling:
		// - If explicit path is provided, use it
		// - If --home flag is used, use home directory
		// - If --global flag is used, use global scope
		// - Otherwise use current directory
		var collectionPath string
		if explicitPath != "" {
			collectionPath = explicitPath
		} else if initUseHomeDir {
			// Keep backward compatibility with --home flag
			collectionPath = "" // Let init handle home directory
		} else if isGlobal {
			// Use global scope path
			path, _ := datapath.ResolveScopedPath(true)
			collectionPath = path
		} else {
			// Default to project scope or current directory
			path, isGlobalScope := datapath.ResolveScopedPath(false)
			if !isGlobalScope {
				// Ensure gitignore is updated for project scope
				if err := datapath.EnsureProjectGitignore(); err != nil {
					// Log but don't fail
					fmt.Printf("Warning: could not update .gitignore: %v\n", err)
				}
			}
			collectionPath = path
		}

		// Call business logic
		result, err := cmdInit.Execute(cmdInit.Options{
			DBPath:     collectionPath,
			UseHomeDir: initUseHomeDir && explicitPath == "",
		})
		if err != nil {
			return err
		}

		// Render output - just use the message field
		return renderToStdout(result.Message)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initUseHomeDir, "home", false, "Create .todos file in home directory instead of current directory")
}
