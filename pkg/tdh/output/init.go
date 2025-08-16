package output

// Import all built-in formatters to ensure they register themselves
import (
	_ "github.com/arthur-debert/tdh/pkg/tdh/output/formatters/json"
	_ "github.com/arthur-debert/tdh/pkg/tdh/output/formatters/term"
)
