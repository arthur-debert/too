package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/arthur-debert/too/pkg/too/store/internal"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate todo collection to pure IDM format",
	Long: `Migrates an existing hierarchical todo collection to the pure IDM format.

This command converts the nested todo structure to a flat collection where
parent-child relationships are managed by IDM rather than embedded in the data.

The original file is backed up with a .backup extension before migration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := dataPath
		if path == "" {
			homeDir, err := getHomeDir()
			if err != nil {
				return err
			}
			path = filepath.Join(homeDir, defaultTodoFile)
		}

		// Check if file exists
		store := internal.NewIDMJSONFileStore(path)
		if !store.Exists() {
			return fmt.Errorf("no todo collection found at %s", path)
		}

		// Load the collection (this will auto-migrate if needed)
		collection, err := store.LoadIDM()
		if err != nil {
			return fmt.Errorf("failed to load collection: %w", err)
		}

		// The LoadIDM method automatically migrates and saves if it detects legacy format
		// So we just need to report success
		fmt.Printf("✓ Successfully migrated todo collection to pure IDM format\n")
		fmt.Printf("  File: %s\n", path)
		fmt.Printf("  Items: %d\n", collection.Count())
		fmt.Printf("\nThe hierarchical structure is now managed by IDM rather than embedded in the data.\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}