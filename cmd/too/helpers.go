package main

import (
	"github.com/arthur-debert/too/pkg/too/commands/datapath"
	"github.com/spf13/cobra"
)

// resolveDataPath resolves the data path from command flags
func resolveDataPath(cmd *cobra.Command) string {
	rawPath, _ := cmd.Flags().GetString("data-path")
	isGlobal, _ := cmd.Flags().GetBool("global")
	return datapath.ResolveCollectionPathWithGlobal(rawPath, isGlobal)
}