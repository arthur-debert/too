package too

import (
	"strings"
	
	"github.com/arthur-debert/too/pkg/too/models"
)

// CommandType represents the category of a command
type CommandType string

const (
	CommandTypeCore  CommandType = "core"
	CommandTypeExtra CommandType = "extras"
	CommandTypeMisc  CommandType = "misc"
)

// AttributeType represents the type of attribute to mutate
type AttributeType string

const (
	AttributeCompletion AttributeType = "completion"
	AttributeText       AttributeType = "text"
	AttributeParent     AttributeType = "parent"
)

// FilterFunc filters a list of todos
type FilterFunc func([]*models.Todo) []*models.Todo

// FilterByStatus returns todos with the given status
func FilterByStatus(status string) FilterFunc {
	return func(todos []*models.Todo) []*models.Todo {
		var filtered []*models.Todo
		for _, todo := range todos {
			if string(todo.GetStatus()) == status {
				filtered = append(filtered, todo)
			}
		}
		return filtered
	}
}

// FilterByQuery returns todos containing the query text
func FilterByQuery(query string) FilterFunc {
	lowerQuery := strings.ToLower(query)
	return func(todos []*models.Todo) []*models.Todo {
		var filtered []*models.Todo
		for _, todo := range todos {
			if strings.Contains(strings.ToLower(todo.Text), lowerQuery) {
				filtered = append(filtered, todo)
			}
		}
		return filtered
	}
}

// FilterAll returns all todos (no filtering)
func FilterAll() FilterFunc {
	return func(todos []*models.Todo) []*models.Todo {
		return todos
	}
}

// FilterPending returns only pending todos
func FilterPending() FilterFunc {
	return FilterByStatus(string(models.StatusPending))
}

// FilterDone returns only done todos
func FilterDone() FilterFunc {
	return FilterByStatus(string(models.StatusDone))
}