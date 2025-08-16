package main

import (
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/spf13/cobra"
)

var swapCmd = &cobra.Command{
	Use:     "swap <source-position> <dest-position>",
	Aliases: []string{"s"},
	Short:   "Swap a todo to a different location",
	Long: `Swap a todo item from one location to another.

The source and destination are specified using position paths:
  - Single number (e.g., "3") refers to a top-level item
  - Dot notation (e.g., "1.2") refers to nested items
  - Empty destination ("") moves to root level

Examples:
  tdh swap 3 1      # Make item 3 a child of item 1
  tdh swap 1.2 ""   # Move item 1.2 to root level
  tdh swap 2.1 3    # Move item 2.1 to be a child of item 3`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		destPath := args[1]

		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic
		result, err := tdh.Swap(sourcePath, destPath, tdh.SwapOptions{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer := output.NewRenderer(nil)
		return renderer.RenderSwap(result)
	},
}

func init() {
	rootCmd.AddCommand(swapCmd)
}
