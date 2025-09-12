package output

import (
	"errors"
	"io"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFormatter is a test formatter
type mockFormatter struct {
	name string
	desc string
}

func (m *mockFormatter) Name() string        { return m.name }
func (m *mockFormatter) Description() string { return m.desc }

func (m *mockFormatter) RenderChange(w io.Writer, result *too.ChangeResult) error         { return nil }
func (m *mockFormatter) RenderMessage(w io.Writer, result *too.MessageResult) error       { return nil }
func (m *mockFormatter) RenderSearch(w io.Writer, result *too.SearchResult) error         { return nil }
func (m *mockFormatter) RenderList(w io.Writer, result *too.ListResult) error             { return nil }
func (m *mockFormatter) RenderFormats(w io.Writer, result *too.ListFormatsResult) error   { return nil }
func (m *mockFormatter) RenderError(w io.Writer, err error) error                         { return nil }

func TestRegistry(t *testing.T) {
	t.Run("Register", func(t *testing.T) {
		// Create a new registry for testing
		reg := &Registry{
			formatters: make(map[string]*FormatterInfo),
		}

		// Test successful registration
		info := &FormatterInfo{
			Info: formatter.Info{
				Name:        "test",
				Description: "Test formatter",
			},
			Factory: func() (Formatter, error) {
				return &mockFormatter{name: "test", desc: "Test formatter"}, nil
			},
		}

		err := reg.Register(info)
		require.NoError(t, err)

		// Test duplicate registration
		err = reg.Register(info)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")

		// Test empty name
		emptyInfo := &FormatterInfo{
			Info: formatter.Info{
				Name:        "",
				Description: "Empty name",
			},
			Factory: info.Factory,
		}
		err = reg.Register(emptyInfo)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("Get", func(t *testing.T) {
		reg := &Registry{
			formatters: make(map[string]*FormatterInfo),
		}

		info := &FormatterInfo{
			Info: formatter.Info{
				Name:        "test",
				Description: "Test formatter",
			},
			Factory: func() (Formatter, error) {
				return &mockFormatter{name: "test", desc: "Test formatter"}, nil
			},
		}

		err := reg.Register(info)
		require.NoError(t, err)

		// Test successful get
		formatter, err := reg.Get("test")
		require.NoError(t, err)
		assert.Equal(t, "test", formatter.Name())
		assert.Equal(t, "Test formatter", formatter.Description())

		// Test get non-existent formatter
		_, err = reg.Get("nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("List", func(t *testing.T) {
		reg := &Registry{
			formatters: make(map[string]*FormatterInfo),
		}

		// Register multiple formatters
		formatters := []string{"alpha", "charlie", "bravo"}
		for _, name := range formatters {
			info := &FormatterInfo{
				Info: formatter.Info{
					Name:        name,
					Description: name + " formatter",
				},
				Factory: func() (Formatter, error) {
					return &mockFormatter{name: name, desc: name + " formatter"}, nil
				},
			}
			err := reg.Register(info)
			require.NoError(t, err)
		}

		// Test list returns sorted names
		names := reg.List()
		assert.Equal(t, []string{"alpha", "bravo", "charlie"}, names)
	})

	t.Run("GetInfo", func(t *testing.T) {
		reg := &Registry{
			formatters: make(map[string]*FormatterInfo),
		}

		// Register multiple formatters
		infos := []*FormatterInfo{
			{
				Info: formatter.Info{
					Name:        "beta",
					Description: "Beta formatter",
				},
				Factory: func() (Formatter, error) {
					return &mockFormatter{}, nil
				},
			},
			{
				Info: formatter.Info{
					Name:        "alpha",
					Description: "Alpha formatter",
				},
				Factory: func() (Formatter, error) {
					return &mockFormatter{}, nil
				},
			},
		}

		for _, info := range infos {
			err := reg.Register(info)
			require.NoError(t, err)
		}

		// Test GetInfo returns sorted infos
		result := reg.GetInfo()
		require.Len(t, result, 2)
		assert.Equal(t, "alpha", result[0].Name)
		assert.Equal(t, "Alpha formatter", result[0].Description)
		assert.Equal(t, "beta", result[1].Name)
		assert.Equal(t, "Beta formatter", result[1].Description)
	})

	t.Run("HasFormatter", func(t *testing.T) {
		reg := &Registry{
			formatters: make(map[string]*FormatterInfo),
		}

		info := &FormatterInfo{
			Info: formatter.Info{
				Name:        "test",
				Description: "Test formatter",
			},
			Factory: func() (Formatter, error) {
				return &mockFormatter{}, nil
			},
		}

		err := reg.Register(info)
		require.NoError(t, err)

		// Test existing formatter
		assert.True(t, reg.HasFormatter("test"))

		// Test non-existent formatter
		assert.False(t, reg.HasFormatter("nonexistent"))
	})

	t.Run("Factory error", func(t *testing.T) {
		reg := &Registry{
			formatters: make(map[string]*FormatterInfo),
		}

		info := &FormatterInfo{
			Info: formatter.Info{
				Name:        "error",
				Description: "Error formatter",
			},
			Factory: func() (Formatter, error) {
				return nil, errors.New("factory error")
			},
		}

		err := reg.Register(info)
		require.NoError(t, err)

		// Test factory error propagation
		_, err = reg.Get("error")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "factory error")
	})
}

func TestGlobalRegistry(t *testing.T) {
	// The global registry should have the term formatter registered
	t.Run("Has term formatter", func(t *testing.T) {
		assert.True(t, HasFormatter("term"))

		formatter, err := Get("term")
		require.NoError(t, err)
		assert.Equal(t, "term", formatter.Name())
		assert.Contains(t, formatter.Description(), "terminal")
	})

	t.Run("List includes term", func(t *testing.T) {
		names := List()
		assert.Contains(t, names, "term")
	})
}
