// Package tdh provides a api facede for the actual api.
// This is intended to facilitate integration with the cli and other tools.
// As not holding any implementation details, this package has no tests .
package tdh

import (
	cmdAdd "github.com/arthur-debert/tdh/pkg/tdh/commands/add"
	cmdClean "github.com/arthur-debert/tdh/pkg/tdh/commands/clean"
	cmdComplete "github.com/arthur-debert/tdh/pkg/tdh/commands/complete"
	cmdInit "github.com/arthur-debert/tdh/pkg/tdh/commands/init"
	cmdList "github.com/arthur-debert/tdh/pkg/tdh/commands/list"
	cmdModify "github.com/arthur-debert/tdh/pkg/tdh/commands/modify"
	cmdReopen "github.com/arthur-debert/tdh/pkg/tdh/commands/reopen"
	cmdReorder "github.com/arthur-debert/tdh/pkg/tdh/commands/reorder"
	cmdSearch "github.com/arthur-debert/tdh/pkg/tdh/commands/search"
	cmdToggle "github.com/arthur-debert/tdh/pkg/tdh/commands/toggle"
)

// Re-export command option types for backward compatibility
type (
	InitOptions     = cmdInit.Options
	AddOptions      = cmdAdd.Options
	ModifyOptions   = cmdModify.Options
	ToggleOptions   = cmdToggle.Options
	CompleteOptions = cmdComplete.Options
	ReopenOptions   = cmdReopen.Options
	CleanOptions    = cmdClean.Options
	ReorderOptions  = cmdReorder.Options
	SearchOptions   = cmdSearch.Options
	ListOptions     = cmdList.Options
)

// Re-export command result types for backward compatibility
type (
	InitResult     = cmdInit.Result
	AddResult      = cmdAdd.Result
	ModifyResult   = cmdModify.Result
	ToggleResult   = cmdToggle.Result
	CompleteResult = cmdComplete.Result
	ReopenResult   = cmdReopen.Result
	CleanResult    = cmdClean.Result
	ReorderResult  = cmdReorder.Result
	SearchResult   = cmdSearch.Result
	ListResult     = cmdList.Result
)

// Init initializes a new todo collection
func Init(opts InitOptions) (*InitResult, error) {
	return cmdInit.Execute(opts)
}

// Add adds a new todo to the collection
func Add(text string, opts AddOptions) (*AddResult, error) {
	return cmdAdd.Execute(text, opts)
}

// Modify modifies the text of an existing todo by position
func Modify(position int, newText string, opts ModifyOptions) (*ModifyResult, error) {
	return cmdModify.Execute(position, newText, opts)
}

// Toggle toggles the status of a todo by position path
func Toggle(positionPath string, opts ToggleOptions) (*ToggleResult, error) {
	return cmdToggle.Execute(positionPath, opts)
}

// Complete marks a todo as complete by position path
func Complete(positionPath string, opts CompleteOptions) (*CompleteResult, error) {
	return cmdComplete.Execute(positionPath, opts)
}

// Reopen marks a todo as pending by position path
func Reopen(positionPath string, opts ReopenOptions) (*ReopenResult, error) {
	return cmdReopen.Execute(positionPath, opts)
}

// Clean removes finished todos from the collection
func Clean(opts CleanOptions) (*CleanResult, error) {
	return cmdClean.Execute(opts)
}

// Reorder reorders todos by sorting them by position and reassigning sequential positions
func Reorder(opts ReorderOptions) (*ReorderResult, error) {
	return cmdReorder.Execute(opts)
}

// Search searches for todos containing the query string
func Search(query string, opts SearchOptions) (*SearchResult, error) {
	return cmdSearch.Execute(query, opts)
}

// List returns todos from the collection with optional filtering
func List(opts ListOptions) (*ListResult, error) {
	return cmdList.Execute(opts)
}
