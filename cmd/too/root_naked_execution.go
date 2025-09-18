package main

import (
	"os"
	"strings"
)

// handleNakedExecution is called when Cobra reports an unknown command
// It checks if we should inject "add" or "list" based on parsed args
func handleNakedExecution() error {
	// Get the args that Cobra parsed (excluding the program name)
	args := os.Args[1:]
	
	// Check for special cases that shouldn't trigger naked execution
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "--version" {
			// Let these pass through normally
			return nil
		}
	}
	
	// Look for any non-flag arguments, properly handling flag values
	hasArgs := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		
		if strings.HasPrefix(arg, "-") {
			// This is a flag - check if it takes a value
			if strings.HasPrefix(arg, "--") {
				// Long flag
				if strings.Contains(arg, "=") {
					// --flag=value format, no need to skip next
					continue
				}
				// Check if this flag takes a value
				switch arg {
				case "--format", "--data-path", "--to":
					if i+1 < len(args) {
						i++ // Skip the value
					}
				}
			} else {
				// Short flag
				switch arg {
				case "-f", "-p":
					if i+1 < len(args) {
						i++ // Skip the value
					}
				}
			}
			continue
		}
		
		// Found a non-flag argument
		hasArgs = true
		break
	}
	
	// Inject the appropriate command
	var newArgs []string
	if hasArgs {
		// Has arguments -> inject "add"
		newArgs = append([]string{os.Args[0], "add"}, os.Args[1:]...)
	} else {
		// No arguments -> inject "list"
		newArgs = append([]string{os.Args[0], "list"}, os.Args[1:]...)
	}
	
	// Update os.Args
	os.Args = newArgs
	
	// Return nil to indicate we should retry
	return nil
}

// isUnknownCommandError checks if the error is due to an unknown command
func isUnknownCommandError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "unknown command") || 
	       strings.Contains(errMsg, "Error: unknown command")
}