// Package store provides the storage adapter for the too todo list application.
//
// The store package wraps the nanostore library to provide todo-specific
// functionality while leveraging nanostore's powerful ID management and
// SQLite-based storage.
//
// # Design Principles
//
// The package follows several key design principles:
//
//   - Uses nanostore for all storage operations and ID management
//   - Provides semi-stable IDs through nanostore's canonical namespace pattern
//   - SQLite-based storage for reliability and performance
//   - Atomic operations with transaction support
//   - Clear error messages with context
//
// # NanoStoreAdapter
//
// The NanoStoreAdapter wraps nanostore.Store and provides todo-specific methods:
//
//   - Add() creates new todos with optional parent
//   - Complete() marks todos as done (moves to 'c' namespace)
//   - Reopen() marks completed todos as pending
//   - Update() modifies todo text
//   - Move() changes todo parent/position
//   - Delete() removes todos (with optional cascade)
//   - List() retrieves todos with filtering
//   - Search() finds todos by text
//
// # Path Resolution
//
// The adapter uses the following path resolution order:
//
//  1. Search current directory and parent directories for .todos.db file
//  2. Check TODO_DB_PATH environment variable
//  3. Fall back to ~/.todos.db in the user's home directory
//
// # ID Management
//
// Nanostore provides semi-stable IDs with canonical namespaces:
//
//   - Pending items: consecutive numbering (1, 2, 3...)
//   - Completed items: 'c' namespace (1.c1, 1.c2...)
//   - IDs remain stable within namespace until structural changes
//
// # Example Usage
//
//	// Create adapter with automatic path resolution
//	adapter, err := store.NewNanoStoreAdapter("")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer adapter.Close()
//
//	// Add a new todo
//	todo, err := adapter.Add("New task", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Complete a todo by position
//	err = adapter.Complete("1")
//	if err != nil {
//	    log.Fatal(err)
//	}
package store
