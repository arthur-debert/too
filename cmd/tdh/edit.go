package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/editor"
	"github.com/spf13/cobra"
)

var editUseEditor bool

var editCmd = &cobra.Command{
	Use:     msgEditUse,
	Aliases: aliasesEdit,
	Short:   msgEditShort,
	Long:    msgEditLong,
	GroupID: "core",
	Args: func(cmd *cobra.Command, args []string) error {
		// Need at least position argument
		if len(args) < 1 {
			return fmt.Errorf("position argument is required")
		}
		// If using editor, we only need position
		if editUseEditor {
			return nil
		}
		// Otherwise, we need position and text
		if len(args) < 2 {
			return fmt.Errorf("both position and text arguments are required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse position
		position, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		var text string
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Handle editor mode
		if editUseEditor {
			// Get current todo to pre-populate editor
			// For now, we'll start with initial content from remaining args if any
			initialContent := ""
			if len(args) > 1 {
				initialContent = strings.Join(args[1:], " ")
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
			// Join remaining arguments as the new text
			text = strings.Join(args[1:], " ")
		}

		// Call business logic
		result, err := tdh.Modify(position, text, tdh.ModifyOptions{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer, err := getRenderer()
		if err != nil {
			return err
		}
		return renderer.RenderModify(result)
	},
}

func init() {
	editCmd.Flags().BoolVarP(&editUseEditor, "editor", "e", false, "open todo in editor for editing")
	rootCmd.AddCommand(editCmd)
}
