package main

import (
	"testing"

	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatFlag(t *testing.T) {
	t.Run("getRenderer with valid format", func(t *testing.T) {
		// Test each registered format
		formats := lipbalm.ListFormats()
		for _, format := range formats {
			formatFlag = format
			renderer, err := getRenderer()
			require.NoError(t, err)
			require.NotNil(t, renderer)
		}
	})

	t.Run("getRenderer with invalid format", func(t *testing.T) {
		formatFlag = "invalid"
		renderer, err := getRenderer()
		assert.Error(t, err)
		assert.Nil(t, renderer)
		assert.Contains(t, err.Error(), "invalid format")
		assert.Contains(t, err.Error(), "Available formats")
	})

	t.Run("default format is term", func(t *testing.T) {
		// Reset to default
		formatFlag = "term"
		renderer, err := getRenderer()
		require.NoError(t, err)
		require.NotNil(t, renderer)
	})
}
