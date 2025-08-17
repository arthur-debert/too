package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: tdh-completion [bash|zsh|fish|powershell]\n")
		os.Exit(1)
	}

	shell := os.Args[1]

	// Create a minimal root command that matches the main tdh command structure
	rootCmd := &cobra.Command{
		Use:   "tdh",
		Short: "A simple command-line todo list manager",
	}

	// Add all the subcommands to ensure proper completion generation
	rootCmd.AddCommand(
		&cobra.Command{Use: "init", Aliases: []string{"i"}, Short: "Initialize a new todo collection"},
		&cobra.Command{Use: "add", Aliases: []string{"a"}, Short: "Add a new todo"},
		&cobra.Command{Use: "modify", Aliases: []string{"m"}, Short: "Modify the text of an existing todo"},
		&cobra.Command{Use: "toggle", Aliases: []string{"t"}, Short: "Toggle the status of a todo"},
		&cobra.Command{Use: "clean", Short: "Remove finished todos"},
		&cobra.Command{Use: "reorder", Aliases: []string{"r"}, Short: "Swap the position of two todos"},
		&cobra.Command{Use: "search", Aliases: []string{"s"}, Short: "Search for todos"},
		&cobra.Command{Use: "list", Short: "List all todos"},
		&cobra.Command{Use: "version", Short: "Print the version number"},
	)

	// Add persistent flags
	rootCmd.PersistentFlags().StringP("collection", "c", "", "path to todo collection")
	rootCmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity")

	// Generate completion based on shell type
	var err error
	switch shell {
	case "bash":
		err = rootCmd.GenBashCompletion(os.Stdout)
	case "zsh":
		err = rootCmd.GenZshCompletion(os.Stdout)
	case "fish":
		err = rootCmd.GenFishCompletion(os.Stdout, true)
	case "powershell":
		err = rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported shell '%s'\n", shell)
		fmt.Fprintf(os.Stderr, "Supported shells: bash, zsh, fish, powershell\n")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating completion: %v\n", err)
		os.Exit(1)
	}
}
