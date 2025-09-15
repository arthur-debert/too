package main

import (
	"testing"

	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatFlag(t *testing.T) {
	t.Run("getRenderEngine with valid format", func(t *testing.T) {
		// Test that we can get a render engine
		engine, err := getRenderEngine()
		require.NoError(t, err)
		require.NotNil(t, engine)
	})

	t.Run("render with valid format", func(t *testing.T) {
		// Test each registered format
		formats := lipbalm.ListFormats()
		for _, format := range formats {
			err := render(&testWriter{}, format, map[string]string{"test": "data"})
			require.NoError(t, err, "format %s should work", format)
		}
	})

	t.Run("render with invalid format", func(t *testing.T) {
		err := render(&testWriter{}, "invalid", map[string]string{"test": "data"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format")
		assert.Contains(t, err.Error(), "Available formats")
	})
}

// testWriter is a simple io.Writer for testing
type testWriter struct{}

func (w *testWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
