// Package too provides a api facade for the actual api.
// This is intended to facilitate integration with the cli and other tools.
package too

import (
	cmdDataPath "github.com/arthur-debert/too/pkg/too/commands/datapath"
	cmdFormats "github.com/arthur-debert/too/pkg/too/commands/formats"
	cmdInit "github.com/arthur-debert/too/pkg/too/commands/init"
)

// Command option types
type (
	// Special commands that aren't unified yet
	InitOptions         = cmdInit.Options
	ShowDataPathOptions = cmdDataPath.Options
	ListFormatsOptions  = cmdFormats.Options
)

// Command result types
type (
	// Special commands
	InitResult         = cmdInit.Result
	ShowDataPathResult = cmdDataPath.Result
	ListFormatsResult  = cmdFormats.Result
	FormatInfo         = cmdFormats.Format
)

// Init initializes a new todo collection
func Init(opts InitOptions) (*InitResult, error) {
	return cmdInit.Execute(opts)
}

// ShowDataPath shows the path to the data file
func ShowDataPath(opts ShowDataPathOptions) (*ShowDataPathResult, error) {
	return cmdDataPath.Execute(opts)
}

// ListFormats returns the list of available output formats
func ListFormats(opts ListFormatsOptions) (*ListFormatsResult, error) {
	return cmdFormats.Execute(opts)
}