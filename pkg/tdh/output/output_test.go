package output

import (
	"bytes"
	"io"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/formatter"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTermFormatter is a mock implementation of the term formatter for testing
type mockTermFormatter struct{}

func (f *mockTermFormatter) Name() string        { return "term" }
func (f *mockTermFormatter) Description() string { return "Mock terminal formatter" }

func (f *mockTermFormatter) RenderAdd(w io.Writer, result *tdh.AddResult) error {
	_, err := w.Write([]byte("mock add output"))
	return err
}

func (f *mockTermFormatter) RenderModify(w io.Writer, result *tdh.ModifyResult) error {
	_, err := w.Write([]byte("mock modify output"))
	return err
}

func (f *mockTermFormatter) RenderInit(w io.Writer, result *tdh.InitResult) error {
	_, err := w.Write([]byte("mock init output"))
	return err
}

func (f *mockTermFormatter) RenderClean(w io.Writer, result *tdh.CleanResult) error {
	_, err := w.Write([]byte("mock clean output"))
	return err
}

func (f *mockTermFormatter) RenderSearch(w io.Writer, result *tdh.SearchResult) error {
	_, err := w.Write([]byte("mock search output"))
	return err
}

func (f *mockTermFormatter) RenderList(w io.Writer, result *tdh.ListResult) error {
	_, err := w.Write([]byte("mock list output"))
	return err
}

func (f *mockTermFormatter) RenderComplete(w io.Writer, results []*tdh.CompleteResult) error {
	_, err := w.Write([]byte("mock complete output"))
	return err
}

func (f *mockTermFormatter) RenderReopen(w io.Writer, results []*tdh.ReopenResult) error {
	_, err := w.Write([]byte("mock reopen output"))
	return err
}

func (f *mockTermFormatter) RenderMove(w io.Writer, result *tdh.MoveResult) error {
	_, err := w.Write([]byte("mock move output"))
	return err
}

func (f *mockTermFormatter) RenderSwap(w io.Writer, result *tdh.SwapResult) error {
	_, err := w.Write([]byte("mock swap output"))
	return err
}

func (f *mockTermFormatter) RenderDataPath(w io.Writer, result *tdh.ShowDataPathResult) error {
	_, err := w.Write([]byte("mock datapath output"))
	return err
}

func (f *mockTermFormatter) RenderFormats(w io.Writer, result *tdh.ListFormatsResult) error {
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
		result := &tdh.AddResult{
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
		result := &tdh.ListResult{
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
