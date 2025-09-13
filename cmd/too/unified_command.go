package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/output"
	"github.com/spf13/cobra"
)

// createUnifiedCommand creates a cobra command from a CommandDef
func createUnifiedCommand(def *too.CommandDef) *cobra.Command {
	cmd := &cobra.Command{
		Use:     def.Name,
		Aliases: def.Aliases,
		Short:   def.Description,
	}

	// Add appropriate Args validator based on command definition
	if def.UsesPositionalArgs {
		if def.AcceptsMultiple {
			cmd.Args = cobra.MinimumNArgs(1)
		} else if def.Name == "move" || def.Name == "edit" {
			// Special cases that need specific arg counts
			if def.Name == "move" {
				cmd.Args = cobra.ExactArgs(2)
			} else {
				cmd.Args = cobra.RangeArgs(1, 2)
			}
		} else {
			cmd.Args = cobra.ExactArgs(1)
		}
	} else {
		cmd.Args = cobra.NoArgs
	}

	// Set the run function
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Build options map from flags
		opts := make(map[string]interface{})
		
		// Get collection path
		rawPath, _ := cmd.Flags().GetString("data-path")
		collectionPath := too.ResolveCollectionPath(rawPath)
		opts["collectionPath"] = collectionPath

		// Common flags
		if def.Name == "list" || def.Name == "search" {
			opts["showDone"], _ = cmd.Flags().GetBool("done")
			opts["showAll"], _ = cmd.Flags().GetBool("all")
		}

		// Command-specific options
		switch def.Name {
		case "add":
			parent, _ := cmd.Flags().GetString("to")
			opts["parent"] = parent
			
			// For add command, handle special text parsing
			if len(args) > 0 {
				// Join all args as the todo text
				text := strings.Join(args, " ")
				// Replace the args with the joined text
				args = []string{text}
			}
		case "search":
			if len(args) > 0 {
				opts["query"] = args[0]
			}
		}

		// Execute the command
		result, err := def.Execute(args, opts)
		if err != nil {
			return err
		}

		// Get renderer
		renderer, err := getRendererForCommand(cmd)
		if err != nil {
			return err
		}

		// Render the result
		return renderer.RenderChange(result)
	}

	// Add command-specific flags
	switch def.Name {
	case "add":
		cmd.Flags().StringP("to", "t", "", "parent todo position")
	case "list", "search":
		cmd.Flags().BoolP("done", "d", false, "show only completed todos")
		cmd.Flags().BoolP("all", "a", false, "show all todos including completed")
	}

	return cmd
}

// getRendererForCommand gets the appropriate renderer based on format flag
func getRendererForCommand(cmd *cobra.Command) (*output.Renderer, error) {
	format, _ := cmd.Flags().GetString("format")
	if format == "" {
		format = "term"
	}

	// Validate the format
	if !output.HasFormatter(format) {
		availableFormats := output.List()
		return nil, fmt.Errorf("invalid format %q. Available formats: %v", format, availableFormats)
	}

	// Create renderer with specified format
	renderer, err := output.NewRendererWithFormat(format, os.Stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}

	return renderer, nil
}