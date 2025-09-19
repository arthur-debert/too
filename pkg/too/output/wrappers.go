package output

import (
	"github.com/arthur-debert/too/pkg/too"
)

// ChangeResultContextual is a wrapper for ChangeResult that triggers the contextual template
type ChangeResultContextual struct {
	*too.ChangeResult
}