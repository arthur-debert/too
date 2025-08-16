package datapath

import (
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// Options holds the options for the datapath command
type Options struct {
	CollectionPath string
}

// Result represents the result of the datapath command
type Result struct {
	Path string
}

// Execute shows the path to the data file
func Execute(opts Options) (*Result, error) {
	// Create a store to get the resolved path
	s := store.NewStore(opts.CollectionPath)

	// Get the path from the store
	path := s.Path()

	return &Result{
		Path: path,
	}, nil
}
