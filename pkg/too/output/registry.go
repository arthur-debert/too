package output

import (
	"fmt"
	"sort"
	"sync"

	"github.com/arthur-debert/too/pkg/too/formatter"
)

// FormatterInfo contains metadata about a formatter with its factory
type FormatterInfo struct {
	formatter.Info
	Factory func() (Formatter, error)
}

// Registry manages available formatters
type Registry struct {
	mu         sync.RWMutex
	formatters map[string]*FormatterInfo
}

// globalRegistry is the singleton registry instance
var globalRegistry = &Registry{
	formatters: make(map[string]*FormatterInfo),
}

// Register adds a formatter to the global registry.
// This should be called from init() functions in formatter packages.
func Register(info *FormatterInfo) error {
	return globalRegistry.Register(info)
}

// Register adds a formatter to the registry
func (r *Registry) Register(info *FormatterInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if info.Name == "" {
		return fmt.Errorf("formatter name cannot be empty")
	}

	if _, exists := r.formatters[info.Name]; exists {
		return fmt.Errorf("formatter %q already registered", info.Name)
	}

	r.formatters[info.Name] = info
	return nil
}

// Get retrieves a formatter by name
func Get(name string) (Formatter, error) {
	return globalRegistry.Get(name)
}

// Get retrieves a formatter by name from the registry
func (r *Registry) Get(name string) (Formatter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.formatters[name]
	if !exists {
		return nil, fmt.Errorf("formatter %q not found", name)
	}

	return info.Factory()
}

// List returns all registered formatter names
func List() []string {
	return globalRegistry.List()
}

// List returns all registered formatter names from the registry
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.formatters))
	for name := range r.formatters {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetInfo returns information about all registered formatters
func GetInfo() []*formatter.Info {
	return globalRegistry.GetInfo()
}

// GetInfo returns information about all registered formatters from the registry
func (r *Registry) GetInfo() []*formatter.Info {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]*formatter.Info, 0, len(r.formatters))
	for _, info := range r.formatters {
		infos = append(infos, &info.Info)
	}

	// Sort by name for consistent output
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos
}

// HasFormatter checks if a formatter is registered
func HasFormatter(name string) bool {
	return globalRegistry.HasFormatter(name)
}

// HasFormatter checks if a formatter is registered in the registry
func (r *Registry) HasFormatter(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.formatters[name]
	return exists
}
