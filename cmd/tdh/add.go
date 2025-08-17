package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/editor"
	"github.com/spf13/cobra"
)

var (
	parentPath string
	useEditor  bool
)

var addCmd = &cobra.Command{
	Use:     msgAddUse,
	Aliases: aliasesAdd,
	Short:   msgAddShort,
	Long:    msgAddLong,
	GroupID: "core",
	Args: func(cmd *cobra.Command, args []string) error {
		// If using editor, we don't need any arguments
		if useEditor {
			return nil
		}
		// Otherwise, we need at least one argument
		if len(args) < 1 {
			return cobra.MinimumNArgs(1)(cmd, args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var text string
		collectionPath, _ := cmd.Flags().GetString("data-path")
		parentPath, _ := cmd.Flags().GetString("to")

		// Handle editor mode
		if useEditor {
			// Get initial content from args if provided
			initialContent := ""
			if len(args) > 0 {
				// Check if first arg is a position path when --to isn't set
				if parentPath == "" && isPositionPath(args[0]) {
					parentPath = args[0]
					// Use remaining args as initial content if any
					if len(args) > 1 {
						initialContent = strings.Join(args[1:], " ")
					}
				} else {
					initialContent = strings.Join(args, " ")
				}
			}

			// Open editor
			editedText, err := editor.OpenInEditor(initialContent)
			if err != nil {
				return err
			}

			// Check if user provided any content
			if editedText == "" {
				return fmt.Errorf("no content provided")
			}

			text = editedText
		} else {
			// Regular mode - parse args as before
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
		}

		// Call business logic
		result, err := tdh.Add(text, tdh.AddOptions{
			CollectionPath: collectionPath,
			ParentPath:     parentPath,
			Mode:           modeFlag,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer, err := getRenderer()
		if err != nil {
			return err
		}
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
	addCmd.Flags().BoolVarP(&useEditor, "editor", "e", false, "open todo in editor for crafting")
	rootCmd.AddCommand(addCmd)
}
