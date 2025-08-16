package formats_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/formats"
)

func TestFormatsCommand(t *testing.T) {
	t.Run("returns the list of available formats", func(t *testing.T) {
		result, err := formats.Execute(formats.Options{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, len(result.Formats), 1)

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
