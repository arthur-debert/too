package formats_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/formats"
	"github.com/arthur-debert/tdh/pkg/tdh/formatter"
)

func TestFormatsCommand(t *testing.T) {
	t.Run("returns the list of available formats", func(t *testing.T) {
		// Set up a mock formatter info function
		originalFunc := formats.GetFormatterInfoFunc
		formats.GetFormatterInfoFunc = func() []*formatter.Info {
			return []*formatter.Info{
				{Name: "json", Description: "JSON output for programmatic consumption"},
				{Name: "term", Description: "Rich terminal output with colors and formatting"},
				{Name: "markdown", Description: "Markdown output for documentation"},
			}
		}
		defer func() {
			formats.GetFormatterInfoFunc = originalFunc
		}()

		result, err := formats.Execute(formats.Options{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3, len(result.Formats))

		// Check that term format is present
		found := false
		for _, format := range result.Formats {
			if format.Name == "term" {
				found = true
				assert.Contains(t, format.Description, "terminal")
				break
			}
		}
		assert.True(t, found, "term format should be registered")
	})
}
