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
	cmdMove "github.com/arthur-debert/tdh/pkg/tdh/commands/move"
	cmdReopen "github.com/arthur-debert/tdh/pkg/tdh/commands/reopen"
	cmdSearch "github.com/arthur-debert/tdh/pkg/tdh/commands/search"
	cmdSwap "github.com/arthur-debert/tdh/pkg/tdh/commands/swap"
)

// Re-export command option types for backward compatibility
type (
	InitOptions     = cmdInit.Options
	AddOptions      = cmdAdd.Options
	ModifyOptions   = cmdModify.Options
	CompleteOptions = cmdComplete.Options
	ReopenOptions   = cmdReopen.Options
	CleanOptions    = cmdClean.Options
	SearchOptions   = cmdSearch.Options
	ListOptions     = cmdList.Options
	MoveOptions     = cmdMove.Options
	SwapOptions     = cmdSwap.Options
)

// Re-export command result types for backward compatibility
type (
	InitResult     = cmdInit.Result
	AddResult      = cmdAdd.Result
	ModifyResult   = cmdModify.Result
	CompleteResult = cmdComplete.Result
	ReopenResult   = cmdReopen.Result
	CleanResult    = cmdClean.Result
	SearchResult   = cmdSearch.Result
	ListResult     = cmdList.Result
	MoveResult     = cmdMove.Result
	SwapResult     = cmdSwap.Result
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

// Search searches for todos containing the query string
func Search(query string, opts SearchOptions) (*SearchResult, error) {
	return cmdSearch.Execute(query, opts)
}

// List returns todos from the collection with optional filtering
func List(opts ListOptions) (*ListResult, error) {
	return cmdList.Execute(opts)
}

// Move moves a todo from one parent to another
func Move(sourcePath string, destParentPath string, opts MoveOptions) (*MoveResult, error) {
	return cmdMove.Execute(sourcePath, destParentPath, opts)
}

// Swap swaps a todo from one parent to another
func Swap(sourcePath string, destParentPath string, opts SwapOptions) (*SwapResult, error) {
	return cmdSwap.Execute(sourcePath, destParentPath, opts)
}
