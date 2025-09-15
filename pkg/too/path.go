package too

import (
	"github.com/arthur-debert/too/pkg/too/commands/datapath"
)

// ResolveCollectionPath is a wrapper that delegates to the datapath package
// to avoid circular imports. This maintains backward compatibility.
func ResolveCollectionPath(explicitPath string) string {
	return datapath.ResolveCollectionPath(explicitPath)
}