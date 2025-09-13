package main

import (
	"fmt"
	"strings"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/editor"
	"github.com/arthur-debert/too/pkg/too/models"
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
		// The position is the first argument
		position := args[0]

		var text string
		rawCollectionPath, _ := cmd.Flags().GetString("data-path")
		collectionPath := too.ResolveCollectionPath(rawCollectionPath)

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
		result, err := too.Modify(position, text, too.ModifyOptions{
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
		
		// Convert to ChangeResult
		changeResult := too.NewChangeResult(
			"placeholder",
			"modified",
			[]*models.IDMTodo{result.Todo},
			result.AllTodos,
			result.TotalCount,
			result.DoneCount,
		)
		
		return renderer.RenderChange(changeResult)
	},
}

func init() {
	editCmd.Flags().BoolVarP(&editUseEditor, "editor", "e", false, "open todo in editor for editing")
	rootCmd.AddCommand(editCmd)
}
