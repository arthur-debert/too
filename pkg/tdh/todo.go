package tdh

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/arthur-debert/tdh/pkg/models"
	"github.com/arthur-debert/tdh/pkg/tdh/printer"
	ct "github.com/daviddengcn/go-colortext"
)

// MakeOutput formats and prints a single todo item to the console.
func MakeOutput(t *models.Todo, useColor bool) {
	var symbole string
	var color ct.Color

	if t.Status == "done" {
		color = ct.Green
		symbole = printer.OkSign
	} else {
		color = ct.Red
		symbole = printer.KoSign
	}

	hashtagReg := regexp.MustCompile(`#[^\\s]*`)
	spaceCount := 6 - len(strconv.FormatInt(t.ID, 10))

	fmt.Print(strings.Repeat(" ", spaceCount), t.ID, " | ")
	if useColor {
		ct.ChangeColor(color, false, ct.None, false)
	}
	fmt.Print(symbole)
	if useColor {
		ct.ResetColor()
	}
	fmt.Print(" ")
	pos := 0
	for _, token := range hashtagReg.FindAllStringIndex(t.Text, -1) {
		fmt.Print(t.Text[pos:token[0]])
		if useColor {
			ct.ChangeColor(ct.Yellow, false, ct.None, false)
		}
		fmt.Print(t.Text[token[0]:token[1]])
		if useColor {
			ct.ResetColor()
		}
		pos = token[1]
	}
	fmt.Println(t.Text[pos:])
}

// RemoveAtIndex removes a todo from a collection at a specific index.
func RemoveAtIndex(c *models.Collection, item int) {
	c.Todos = append(c.Todos[:item], c.Todos[item+1:]...)
}

// ListPendingTodos filters the collection to only show pending todos.
func ListPendingTodos(c *models.Collection) {
	for i := len(c.Todos) - 1; i >= 0; i-- {
		if c.Todos[i].Status != "pending" {
			RemoveAtIndex(c, i)
		}
	}
}

// ListDoneTodos filters the collection to only show done todos.
func ListDoneTodos(c *models.Collection) {
	for i := len(c.Todos) - 1; i >= 0; i-- {
		if c.Todos[i].Status != "done" {
			RemoveAtIndex(c, i)
		}
	}
}

// Find finds a todo by its ID in a collection.
func Find(c *models.Collection, id int) (*models.Todo, error) {
	id64 := int64(id)
	for _, todo := range c.Todos {
		if id64 == todo.ID {
			return todo, nil
		}
	}
	return nil, errors.New("The todo with the id " + strconv.FormatInt(id64, 10) + " was not found.")
}

// RemoveFinishedTodos removes all done todos from a collection.
func RemoveFinishedTodos(c *models.Collection) int {
	var activeTodos []*models.Todo
	for _, todo := range c.Todos {
		if todo.Status != "done" {
			activeTodos = append(activeTodos, todo)
		}
	}
	c.Todos = activeTodos
	return len(activeTodos)
}

// Swap swaps the position of two todos in a collection by their IDs.
func Swap(c *models.Collection, idA, idB int) error {
	var positionA, positionB = -1, -1
	idA64 := int64(idA)
	idB64 := int64(idB)

	for i, todo := range c.Todos {
		if todo.ID == idA64 {
			positionA = i
		}
		if todo.ID == idB64 {
			positionB = i
		}
	}

	if positionA == -1 || positionB == -1 {
		return errors.New("one or both todos not found")
	}

	c.Todos[positionA], c.Todos[positionB] = c.Todos[positionB], c.Todos[positionA]
	c.Todos[positionA].ID, c.Todos[positionB].ID = c.Todos[positionB].ID, c.Todos[positionA].ID
	return nil
}
