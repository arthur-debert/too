package main

import (
	"github.com/arthur-debert/too/pkg/too/commands/formats"
	"github.com/spf13/cobra"
)

var formatsCmd = &cobra.Command{
	Use:     "formats",
	Short:   "List available output formats",
	Long:    "Display a list of the available formats for command output.",
	GroupID: "misc",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Call business logic
		result, err := formats.Execute(formats.Options{})
		if err != nil {
			return err
		}

		// Render the result
		renderer, err := getRenderer()
		if err != nil {
			return err
		}
		return renderer.RenderFormats(result)
	},
}

func init() {
	rootCmd.AddCommand(formatsCmd)
}
