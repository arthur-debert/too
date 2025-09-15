package formats_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arthur-debert/too/pkg/too/commands/formats"
)

func TestFormatsCommand(t *testing.T) {
	t.Run("returns the list of available formats", func(t *testing.T) {
		result, err := formats.Execute(formats.Options{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, len(result.Formats), 0, "should have at least one format")

		// Check that standard formats are present
		formatNames := make(map[string]string)
		for _, format := range result.Formats {
			formatNames[format.Name] = format.Description
		}

		// Verify expected formats
		assert.Contains(t, formatNames, "json")
		assert.Contains(t, formatNames, "term")
		assert.Contains(t, formatNames, "yaml")
		assert.Contains(t, formatNames, "markdown")

		// Check that term format has appropriate description
		if termDesc, exists := formatNames["term"]; exists {
			assert.Contains(t, termDesc, "terminal")
		}
	})
}
