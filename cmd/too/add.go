package main

import (
	"fmt"
	"strings"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/commands/datapath"
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
		collectionPath := resolveDataPath(cmd)
		parentPath, _ := cmd.Flags().GetString("to")
		
		// Ensure gitignore is updated for project scope
		if err := datapath.EnsureProjectGitignore(); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: could not update .gitignore: %v\n", err)
		}

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
			// Regular mode - parse args
			// If --to wasn't explicitly set and we have at least 2 args
			if parentPath == "" && len(args) >= 2 {
				// Check LAST arg for position path pattern
				lastArg := args[len(args)-1]
				if parser.IsPositionPath(lastArg) {
					// Last arg is position path, everything else is text
					parentPath = lastArg
					text = strings.Join(args[:len(args)-1], " ")
				} else {
					// No position path found, all args are text
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
			var changeResults []*too.ChangeResult

			// Helper function to add todos recursively
			var addTodoWithChildren func(todo *parser.TodoItem, parentPath string) error
			addTodoWithChildren = func(todo *parser.TodoItem, parentPath string) error {
				// Add the current todo
				opts := map[string]interface{}{
					"collectionPath": collectionPath,
					"parent":         parentPath,
				}
				result, err := too.ExecuteUnifiedCommand("add", []string{todo.Text}, opts)
				if err != nil {
					return fmt.Errorf("failed to add todo '%s': %w", todo.Text, err)
				}
				changeResults = append(changeResults, result)

				// Add children with this todo as parent
				if len(todo.Children) > 0 && len(result.AffectedTodos) > 0 {
					// Use the position path of the newly created todo
					newParentPath := result.AffectedTodos[0].PositionPath

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
			// Render the last result which has the final state
			if len(changeResults) > 0 {
				lastResult := changeResults[len(changeResults)-1]
				// Update message to reflect total added
				lastResult.Message = fmt.Sprintf("Added %d todos", len(changeResults))
				
				if err := renderToStdout(lastResult); err != nil {
					return err
				}
			}

			return nil
		}

		// Call business logic using unified command
		opts := map[string]interface{}{
			"collectionPath": collectionPath,
			"parent":         parentPath,
		}
		result, err := too.ExecuteUnifiedCommand("add", []string{text}, opts)
		if err != nil {
			return err
		}

		// Render output
		return renderToStdout(result)
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
