package tdh

import (
	cmdAdd "github.com/arthur-debert/tdh/pkg/tdh/commands/add"
	cmdClean "github.com/arthur-debert/tdh/pkg/tdh/commands/clean"
	cmdInit "github.com/arthur-debert/tdh/pkg/tdh/commands/init"
	cmdList "github.com/arthur-debert/tdh/pkg/tdh/commands/list"
	cmdModify "github.com/arthur-debert/tdh/pkg/tdh/commands/modify"
	cmdReorder "github.com/arthur-debert/tdh/pkg/tdh/commands/reorder"
	cmdSearch "github.com/arthur-debert/tdh/pkg/tdh/commands/search"
	cmdToggle "github.com/arthur-debert/tdh/pkg/tdh/commands/toggle"
)

// Re-export command option types for backward compatibility
type (
	InitOptions    = cmdInit.Options
	AddOptions     = cmdAdd.Options
	ModifyOptions  = cmdModify.Options
	ToggleOptions  = cmdToggle.Options
	CleanOptions   = cmdClean.Options
	ReorderOptions = cmdReorder.Options
	SearchOptions  = cmdSearch.Options
	ListOptions    = cmdList.Options
)

// Re-export command result types for backward compatibility
type (
	InitResult    = cmdInit.Result
	AddResult     = cmdAdd.Result
	ModifyResult  = cmdModify.Result
	ToggleResult  = cmdToggle.Result
	CleanResult   = cmdClean.Result
	ReorderResult = cmdReorder.Result
	SearchResult  = cmdSearch.Result
	ListResult    = cmdList.Result
)

// Init initializes a new todo collection
func Init(opts InitOptions) (*InitResult, error) {
	return cmdInit.Execute(opts)
}

// Add adds a new todo to the collection
func Add(text string, opts AddOptions) (*AddResult, error) {
	return cmdAdd.Execute(text, opts)
}

// Modify modifies the text of an existing todo
func Modify(id int, newText string, opts ModifyOptions) (*ModifyResult, error) {
	return cmdModify.Execute(id, newText, opts)
}

// Toggle toggles the status of a todo
func Toggle(id int, opts ToggleOptions) (*ToggleResult, error) {
	return cmdToggle.Execute(id, opts)
}

// Clean removes finished todos from the collection
func Clean(opts CleanOptions) (*CleanResult, error) {
	return cmdClean.Execute(opts)
}

// Reorder swaps the position of two todos
func Reorder(idA, idB int, opts ReorderOptions) (*ReorderResult, error) {
	return cmdReorder.Execute(idA, idB, opts)
}

// Search searches for todos containing the query string
func Search(query string, opts SearchOptions) (*SearchResult, error) {
	return cmdSearch.Execute(query, opts)
}

// List returns todos from the collection with optional filtering
func List(opts ListOptions) (*ListResult, error) {
	return cmdList.Execute(opts)
}
