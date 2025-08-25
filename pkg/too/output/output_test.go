package output

import (
	"bytes"
	"io"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/formatter"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTermFormatter is a mock implementation of the term formatter for testing
type mockTermFormatter struct{}

func (f *mockTermFormatter) Name() string        { return "term" }
func (f *mockTermFormatter) Description() string { return "Mock terminal formatter" }

func (f *mockTermFormatter) RenderAdd(w io.Writer, result *too.AddResult) error {
	_, err := w.Write([]byte("mock add output"))
	return err
}

func (f *mockTermFormatter) RenderModify(w io.Writer, result *too.ModifyResult) error {
	_, err := w.Write([]byte("mock modify output"))
	return err
}

func (f *mockTermFormatter) RenderInit(w io.Writer, result *too.InitResult) error {
	_, err := w.Write([]byte("mock init output"))
	return err
}

func (f *mockTermFormatter) RenderClean(w io.Writer, result *too.CleanResult) error {
	_, err := w.Write([]byte("mock clean output"))
	return err
}

func (f *mockTermFormatter) RenderSearch(w io.Writer, result *too.SearchResult) error {
	_, err := w.Write([]byte("mock search output"))
	return err
}

func (f *mockTermFormatter) RenderList(w io.Writer, result *too.ListResult) error {
	_, err := w.Write([]byte("mock list output"))
	return err
}

func (f *mockTermFormatter) RenderComplete(w io.Writer, results []*too.CompleteResult) error {
	_, err := w.Write([]byte("mock complete output"))
	return err
}

func (f *mockTermFormatter) RenderReopen(w io.Writer, results []*too.ReopenResult) error {
	_, err := w.Write([]byte("mock reopen output"))
	return err
}

func (f *mockTermFormatter) RenderMove(w io.Writer, result *too.MoveResult) error {
	_, err := w.Write([]byte("mock move output"))
	return err
}

func (f *mockTermFormatter) RenderSwap(w io.Writer, result *too.SwapResult) error {
	_, err := w.Write([]byte("mock swap output"))
	return err
}

func (f *mockTermFormatter) RenderDataPath(w io.Writer, result *too.ShowDataPathResult) error {
	_, err := w.Write([]byte("mock datapath output"))
	return err
}

func (f *mockTermFormatter) RenderFormats(w io.Writer, result *too.ListFormatsResult) error {
	_, err := w.Write([]byte("mock formats output"))
	return err
}

func (f *mockTermFormatter) RenderError(w io.Writer, err error) error {
	_, writeErr := w.Write([]byte("mock error output"))
	return writeErr
}

func init() {
	// Register mock term formatter for tests
	if err := Register(&FormatterInfo{
		Info: formatter.Info{
			Name:        "term",
			Description: "Mock terminal formatter",
		},
		Factory: func() (Formatter, error) {
			return &mockTermFormatter{}, nil
		},
	}); err != nil {
		panic("failed to register mock formatter: " + err.Error())
	}
}

func TestNewRenderer(t *testing.T) {
	t.Run("Default renderer uses term formatter", func(t *testing.T) {
		renderer := NewRenderer(nil)
		require.NotNil(t, renderer)
		assert.NotNil(t, renderer.formatter)
		assert.Equal(t, "term", renderer.formatter.Name())
	})

	t.Run("Custom writer is used", func(t *testing.T) {
		buf := &bytes.Buffer{}
		renderer := NewRenderer(buf)
		require.NotNil(t, renderer)
		assert.Equal(t, buf, renderer.writer)
	})
}

func TestNewRendererWithFormat(t *testing.T) {
	t.Run("Valid format", func(t *testing.T) {
		renderer, err := NewRendererWithFormat("term", nil)
		require.NoError(t, err)
		require.NotNil(t, renderer)
		assert.Equal(t, "term", renderer.formatter.Name())
	})

	t.Run("Invalid format returns error", func(t *testing.T) {
		_, err := NewRendererWithFormat("invalid", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get formatter")
	})
}

func TestRendererMethods(t *testing.T) {
	// Create a renderer with buffer to capture output
	buf := &bytes.Buffer{}
	renderer := NewRenderer(buf)

	t.Run("RenderAdd", func(t *testing.T) {
		result := &too.AddResult{
			Todo: &models.Todo{
				Position: 1,
				Text:     "Test todo",
				Status:   models.StatusPending,
			},
		}
		err := renderer.RenderAdd(result)
		require.NoError(t, err)
		// Check that something was written
		assert.NotEmpty(t, buf.String())
	})

	t.Run("RenderList", func(t *testing.T) {
		buf.Reset()
		result := &too.ListResult{
			Todos: []*models.Todo{
				{
					Position: 1,
					Text:     "First todo",
					Status:   models.StatusPending,
				},
				{
					Position: 2,
					Text:     "Second todo",
					Status:   models.StatusDone,
				},
			},
			TotalCount: 2,
			DoneCount:  1,
		}
		err := renderer.RenderList(result)
		require.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})
}
