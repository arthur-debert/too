package main

import (
	"regexp"
	"strings"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/spf13/cobra"
)

var (
	parentPath string
)

var addCmd = &cobra.Command{
	Use:     msgAddUse,
	Aliases: aliasesAdd,
	Short:   msgAddShort,
	Long:    msgAddLong,
	GroupID: "core",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if first argument is a position path (shortcut for --to)
		var text string
		collectionPath, _ := cmd.Flags().GetString("data-path")
		parentPath, _ := cmd.Flags().GetString("to")

		// If --to wasn't explicitly set and we have at least 2 args
		if parentPath == "" && len(args) >= 2 {
			// Check if first arg matches position path pattern (e.g., "1", "1.2", "1.2.3")
			if isPositionPath(args[0]) {
				parentPath = args[0]
				text = strings.Join(args[1:], " ")
			} else {
				// Normal case: all args are the todo text
				text = strings.Join(args, " ")
			}
		} else {
			// Normal case: all args are the todo text
			text = strings.Join(args, " ")
		}

		// Call business logic
		result, err := tdh.Add(text, tdh.AddOptions{
			CollectionPath: collectionPath,
			ParentPath:     parentPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer := output.NewRenderer(nil)
		return renderer.RenderAdd(result)
	},
}

// isPositionPath checks if a string matches the position path pattern (e.g., "1", "1.2", "1.2.3")
func isPositionPath(s string) bool {
	// Pattern: one or more digits, optionally followed by dot and more digits
	// Examples: "1", "12", "1.2", "1.2.3", "12.34.56"
	pattern := `^\d+(\.\d+)*$`
	matched, _ := regexp.MatchString(pattern, strings.TrimSpace(s))
	return matched
}

func init() {
	addCmd.Flags().StringVar(&parentPath, "to", "", "parent todo position path (e.g., \"1.2\")")
	rootCmd.AddCommand(addCmd)
}
