package parser

import (
	"strings"
)

// TodoItem represents a parsed todo with its content and nested items
type TodoItem struct {
	Text     string
	Children []*TodoItem
	Level    int // Indentation level (0 = root)
}

// ParseOptions configures the parser behavior
type ParseOptions struct {
	// TabWidth defines how many spaces a tab represents
	TabWidth int
	// MinIndent defines the minimum spaces needed for a level
	MinIndent int
}

// DefaultParseOptions returns the default parsing options
func DefaultParseOptions() ParseOptions {
	return ParseOptions{
		TabWidth:  4,
		MinIndent: 2,
	}
}

// ParseMultipleTodos parses text containing multiple todos with support for nesting
func ParseMultipleTodos(text string, opts ParseOptions) []*TodoItem {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return nil
	}

	// Normalize tabs to spaces
	for i, line := range lines {
		lines[i] = strings.ReplaceAll(line, "\t", strings.Repeat(" ", opts.TabWidth))
	}

	todos := make([]*TodoItem, 0)
	stack := make([]*TodoItem, 0) // Stack to track parent todos at each level

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Check if this line starts a new todo
		indent, isTodo, content := parseTodoLine(line)
		if !isTodo {
			// If not a todo line, it might be continuation of previous todo
			if len(stack) > 0 {
				// Add to the most recent todo's text
				currentTodo := stack[len(stack)-1]
				if strings.TrimSpace(line) != "" {
					currentTodo.Text += "\n" + strings.TrimSpace(line)
				}
			}
			i++
			continue
		}

		// Calculate the level based on indentation
		level := indent / opts.MinIndent

		// Create new todo item
		todo := &TodoItem{
			Text:     content,
			Children: make([]*TodoItem, 0),
			Level:    level,
		}

		// Find the correct parent based on level
		if level == 0 {
			// Root level todo
			todos = append(todos, todo)
			stack = []*TodoItem{todo}
		} else {
			// Adjust stack to correct size
			for len(stack) > level {
				stack = stack[:len(stack)-1]
			}

			// Ensure we have enough stack depth
			for len(stack) < level {
				// Create placeholder if needed
				if len(stack) == 0 {
					// Can't have nested todo without parent, treat as root
					level = 0
					todo.Level = 0
					todos = append(todos, todo)
					stack = []*TodoItem{todo}
					break
				}
				stack = append(stack, stack[len(stack)-1])
			}

			if len(stack) > 0 && level > 0 && level != todo.Level {
				// Skip if we already added as root
				i++
				continue
			}

			if len(stack) > 0 && level > 0 {
				parent := stack[level-1]
				parent.Children = append(parent.Children, todo)

				// Update or extend stack
				if len(stack) == level {
					stack = append(stack, todo)
				} else {
					stack[level] = todo
				}
			}
		}

		i++
	}

	return todos
}

// parseTodoLine checks if a line is a todo and extracts its components
func parseTodoLine(line string) (indent int, isTodo bool, content string) {
	// Count leading spaces
	indent = 0
	for i, ch := range line {
		if ch == ' ' {
			indent++
		} else {
			line = line[i:]
			break
		}
	}

	// Check for todo markers (- or *)
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "- ") {
		return indent, true, strings.TrimPrefix(line, "- ")
	} else if strings.HasPrefix(line, "* ") {
		return indent, true, strings.TrimPrefix(line, "* ")
	}

	return indent, false, ""
}

// Flatten converts a nested todo structure into a flat list with proper text formatting
func Flatten(todos []*TodoItem) []string {
	result := make([]string, 0)
	flattenRecursive(todos, &result, "")
	return result
}

func flattenRecursive(todos []*TodoItem, result *[]string, prefix string) {
	for i, todo := range todos {
		text := todo.Text
		if prefix != "" {
			text = prefix + "." + text
		}
		*result = append(*result, text)

		if len(todo.Children) > 0 {
			childPrefix := prefix
			if childPrefix == "" {
				childPrefix = string(rune('0' + i + 1))
			} else {
				childPrefix = childPrefix + "." + string(rune('0'+i+1))
			}
			flattenRecursive(todo.Children, result, childPrefix)
		}
	}
}
