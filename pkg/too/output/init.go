package output

import (
	"github.com/arthur-debert/too/pkg/too/commands/formats"
)

func init() {
	// Set the function to retrieve formatter info to avoid import cycles
	formats.GetFormatterInfoFunc = GetInfo
}