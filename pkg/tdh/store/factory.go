package store

import (
	"errors"
	"os"
	"path"
	"path/filepath"
)

var (
	errIsNotAFile          = errors.New("the database path is not a file")
	errLocalDbFileNotFound = errors.New("the local .todos file was not found")
	cachedDBPath           = ""
)

// tryDir checks if a .todos file exists in the given directory
func tryDir(dir string) (string, error) {
	dbPath := path.Join(dir, ".todos")
	fi, err := os.Stat(dbPath)
	if err != nil {
		return "", err
	}

	if fi.IsDir() {
		return "", errIsNotAFile
	}

	return dbPath, nil
}

// tryCwdAndParentFolders searches for .todos file in current directory and parent directories
func tryCwdAndParentFolders() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		filePath, err := tryDir(cwd)
		if err == nil {
			return filePath, err
		}

		if len(cwd) == 1 {
			break
		}

		cwd = path.Dir(cwd)
	}

	return "", errLocalDbFileNotFound
}

// tryEnv checks the TODO_DB_PATH environment variable
func tryEnv() (string, error) {
	envPath := os.Getenv("TODO_DB_PATH")
	if envPath != "" {
		return envPath, nil
	}
	return "", errors.New("TODO_DB_PATH not set")
}

// calculateDBPath determines the database path using the original logic:
// 1. Search current directory and parents for .todos file
// 2. Check TODO_DB_PATH environment variable
// 3. Use default home directory path
func calculateDBPath() string {
	// Try cached path first
	if cachedDBPath != "" {
		return cachedDBPath
	}

	// Try current directory and parent folders
	dbPath, err := tryCwdAndParentFolders()
	if err == nil {
		cachedDBPath = dbPath
		return dbPath
	}

	// Try environment variable
	dbPath, err = tryEnv()
	if err == nil {
		cachedDBPath = dbPath
		return dbPath
	}

	// Fall back to default home directory path
	home, err := os.UserHomeDir()
	if err != nil {
		// Last resort fallback if home dir is not available
		cachedDBPath = ".todos.json"
		return cachedDBPath
	}

	cachedDBPath = filepath.Join(home, ".todos.json")
	return cachedDBPath
}

// NewStore creates a new store based on the provided path.
// If path is empty, it uses the path resolution logic from the original db.go:
// 1. Search current directory and parents for .todos file
// 2. Check TODO_DB_PATH environment variable
// 3. Use default home directory path (~/.todos.json)
func NewStore(path string) Store {
	if path == "" {
		path = calculateDBPath()
	}
	return NewJSONFileStore(path)
}
