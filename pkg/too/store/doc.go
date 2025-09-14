// Package store provides storage abstractions for the too todo list application.
//
// The store package implements a clean separation between business logic and
// persistence concerns, allowing the too application to work with different
// storage backends without changing the core logic.
//
// # Design Principles
//
// The package follows several key design principles:
//
//   - Storage operations are abstracted behind the Store interface
//   - All saves are atomic using a write-and-rename pattern to prevent data corruption
//   - Updates are transactional with automatic rollback on error
//   - Path resolution maintains backward compatibility with the original too
//   - Error messages provide clear context about what failed
//
// # Store Interface
//
// The Store interface defines the contract for all storage implementations:
//
//   - Load() retrieves the todo collection from storage
//   - Save() persists the todo collection to storage
//   - Exists() checks if the storage exists
//   - Update() performs transactional updates with rollback safety
//   - Path() returns the storage location
//
// # Implementation
//
// The current implementation uses NanoStoreAdapter which wraps the nanostore
// library for SQLite-based storage with dynamic ID generation
//
// # Path Resolution
//
// The NanoStoreAdapter uses the following path resolution order:
//
//  1. Search current directory and parent directories for .todos.db file
//  2. Check TODO_DB_PATH environment variable
//  3. Fall back to ~/.todos.db in the user's home directory
//
// This maintains compatibility with existing too installations while
// allowing users to override the default location.
//
// # Example Usage
//
//	// Create a store with automatic path resolution
//	store := store.NewStore("")
//
//	// Load existing todos
//	collection, err := store.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Update todos transactionally
//	err = store.Update(func(c *models.Collection) error {
//	    c.CreateTodo("New task")
//	    return nil
//	})
package store

// RootScope is the special scope identifier for the root of the todo tree.
const RootScope = "root"
