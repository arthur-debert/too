package main

import (
	"fmt"
	"strings"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/editor"
	"github.com/arthur-debert/too/pkg/too/parser"
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
		rawCollectionPath, _ := cmd.Flags().GetString("data-path")
		collectionPath := too.ResolveCollectionPath(rawCollectionPath)
		parentPath, _ := cmd.Flags().GetString("to")

		// Handle editor mode
		if useEditor {
			// Get initial content from args if provided
			initialContent := ""
			if len(args) > 0 {
				// Check if first arg is a position path when --to isn't set
				if parentPath == "" && parser.IsPositionPath(args[0]) {
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
				if parser.IsPositionPath(args[0]) {
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

		// Check if text contains multiple todos (bullet points)
		if containsBulletPoints(text) {
			// Parse multiple todos
			todos := parser.ParseMultipleTodos(text, parser.DefaultParseOptions())
			if len(todos) == 0 {
				return fmt.Errorf("no todos found in input")
			}

			// For now, we'll add them one by one
			// TODO: Create a batch add function for efficiency
			results := make([]*too.AddResult, 0)

			// Helper function to add todos recursively
			var addTodoWithChildren func(todo *parser.TodoItem, parentPath string) error
			addTodoWithChildren = func(todo *parser.TodoItem, parentPath string) error {
				// Add the current todo
				result, err := too.Add(todo.Text, too.AddOptions{
					CollectionPath: collectionPath,
					ParentPath:     parentPath,
					Mode:           modeFlag,
				})
				if err != nil {
					return fmt.Errorf("failed to add todo '%s': %w", todo.Text, err)
				}
				results = append(results, result)

				// Add children with this todo as parent
				if len(todo.Children) > 0 {
					// Use the position path of the newly created todo
					newParentPath := result.PositionPath

					for _, child := range todo.Children {
						if err := addTodoWithChildren(child, newParentPath); err != nil {
							return err
						}
					}
				}

				return nil
			}

			// Add all root-level todos
			for _, todo := range todos {
				if err := addTodoWithChildren(todo, parentPath); err != nil {
					return err
				}
			}

			// Render all results
			renderer, err := getRenderer()
			if err != nil {
				return err
			}

			// For now, render each result individually
			// TODO: Create a batch render method
			for _, result := range results {
				if err := renderer.RenderAdd(result); err != nil {
					return err
				}
			}

			return nil
		}

		// Call business logic
		result, err := too.Add(text, too.AddOptions{
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

// containsBulletPoints checks if text contains markdown-style bullet points
func containsBulletPoints(text string) bool {
	// Check for lines starting with - or * (with optional leading whitespace)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			return true
		}
	}
	return false
}

func init() {
	addCmd.Flags().StringVar(&parentPath, "to", "", "parent todo position path (e.g., \"1.2\")")
	addCmd.Flags().BoolVarP(&useEditor, "editor", "e", false, "open todo in editor for crafting")
	rootCmd.AddCommand(addCmd)
}
