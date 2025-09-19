package datapath

import (
	"os"
	"path/filepath"
	"strings"
	
	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/arthur-debert/too/pkg/too/scope"
)

// ResolveCollectionPath resolves the collection file path using the following order:
// 1. If explicitPath is provided, use it as-is (supports tilde expansion)
// 2. Check TODO_DB_PATH environment variable
// 3. Use scope-based resolution (project vs global)
func ResolveCollectionPath(explicitPath string) string {
	return ResolveCollectionPathWithGlobal(explicitPath, false)
}

// ResolveCollectionPathWithGlobal resolves the collection file path with global flag support
func ResolveCollectionPathWithGlobal(explicitPath string, forceGlobal bool) string {
	if explicitPath != "" {
		// Handle tilde expansion
		if strings.HasPrefix(explicitPath, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				return filepath.Join(home, explicitPath[2:])
			}
		}
		return explicitPath
	}

	// Check TODO_DB_PATH environment variable
	if envPath := os.Getenv("TODO_DB_PATH"); envPath != "" {
		return envPath
	}

	// Use scoped path resolution
	path, _ := ResolveScopedPath(forceGlobal)
	return path
}

// ResolveScopedPath determines the appropriate storage path based on scope
// Returns the path and whether it's global scope
func ResolveScopedPath(forceGlobal bool) (string, bool) {
	resolver := scope.NewResolver(forceGlobal)
	
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}
	
	scopeInfo, err := resolver.Resolve(currentDir)
	if err != nil {
		// Fallback to current directory
		return ".todos.json", false
	}
	
	// Ensure parent directory exists for global path
	if scopeInfo.IsGlobal {
		parentDir := filepath.Dir(scopeInfo.Path)
		_ = os.MkdirAll(parentDir, 0755) // Best effort - ignore errors
	}
	
	return scopeInfo.Path, scopeInfo.IsGlobal
}

// EnsureProjectGitignore ensures .todos.json is in .gitignore for project scope
func EnsureProjectGitignore() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil // Silently ignore errors
	}
	
	resolver := scope.NewResolver(false)
	scopeInfo, err := resolver.Resolve(currentDir)
	if err != nil || scopeInfo.IsGlobal || scopeInfo.GitRoot == "" {
		return nil // Not in a git repo or using global scope
	}
	
	return scope.EnsureGitignore(scopeInfo.GitRoot)
}

// Options holds the options for the datapath command
type Options struct {
	CollectionPath string
}

// Execute shows the path to the data file
func Execute(opts Options) (*lipbalm.Message, error) {
	// Use the unified path resolution function
	storePath := ResolveCollectionPath(opts.CollectionPath)

	// Get the absolute path
	absPath, err := filepath.Abs(storePath)
	if err != nil {
		// If we can't get absolute path, just return the path as is
		absPath = storePath
	}

	// Return the path as a plain message for proper rendering
	return lipbalm.NewPlainMessage(absPath), nil
}
