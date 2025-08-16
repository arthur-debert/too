package output

import (
	"github.com/arthur-debert/tdh/pkg/tdh/commands/formats"
)

func init() {
	// Set the function to retrieve formatter info to avoid import cycles
	formats.GetFormatterInfoFunc = GetInfo
}