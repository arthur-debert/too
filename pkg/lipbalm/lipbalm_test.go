package lipbalm_test

import (
	"bytes"
	"embed"
	"io"
	"strings"
	"testing"
	"text/template"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/arthur-debert/too/pkg/lipbalm"
)

//go:embed testdata/*.tmpl
var testTemplateFS embed.FS

var testStyles = lipbalm.StyleMap{
	"title": lipgloss.NewStyle().Bold(true),
	"date":  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	"body":  lipgloss.NewStyle().Italic(true),
}

func TestRender(t *testing.T) {
	// a buffer to render to
	var buf bytes.Buffer
	renderer := lipgloss.NewRenderer(&buf)
	lipbalm.SetDefaultRenderer(renderer)

	t.Run("go template expansion", func(t *testing.T) {
		renderer.SetColorProfile(termenv.TrueColor)
		template := `<title>{{.Title}}</title>`
		data := struct{ Title string }{Title: "My Title"}
		expected := testStyles["title"].Render("My Title")
		actual, err := lipbalm.Render(template, data, testStyles)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("invalid go template", func(t *testing.T) {
		renderer.SetColorProfile(termenv.TrueColor)
		template := `<title>{{.Title</title>`
		_, err := lipbalm.Render(template, nil, testStyles)
		assert.Error(t, err)
	})

	t.Run("with custom template functions", func(t *testing.T) {
		renderer.SetColorProfile(termenv.TrueColor)
		funcs := template.FuncMap{
			"upper": strings.ToUpper,
			"repeat": func(s string, n int) string { return strings.Repeat(s, n) },
		}
		template := `<title>{{upper .Title}}</title> <date>{{repeat "*" 3}}</date>`
		data := struct{ Title string }{Title: "hello"}
		expected := testStyles["title"].Render("HELLO") + " " + testStyles["date"].Render("***")
		actual, err := lipbalm.Render(template, data, testStyles, funcs)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestTemplateManager(t *testing.T) {
	// a buffer to render to
	var buf bytes.Buffer
	renderer := lipgloss.NewRenderer(&buf)
	lipbalm.SetDefaultRenderer(renderer)

	t.Run("basic template management", func(t *testing.T) {
		renderer.SetColorProfile(termenv.TrueColor)
		
		funcs := template.FuncMap{
			"upper": strings.ToUpper,
		}
		tm := lipbalm.NewTemplateManager(testStyles, funcs)
		
		// Add a template manually
		tm.AddTemplate("test", `<title>{{upper .Title}}</title>`)
		
		// Render the template
		data := struct{ Title string }{Title: "hello"}
		expected := testStyles["title"].Render("HELLO")
		actual, err := tm.RenderTemplate("test", data)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
		
		// Check template listing
		templates := tm.ListTemplates()
		assert.Contains(t, templates, "test")
	})

	t.Run("template not found", func(t *testing.T) {
		tm := lipbalm.NewTemplateManager(testStyles, nil)
		_, err := tm.RenderTemplate("nonexistent", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template 'nonexistent' not found")
	})

	t.Run("load templates from embedded filesystem", func(t *testing.T) {
		renderer.SetColorProfile(termenv.TrueColor)
		
		tm := lipbalm.NewTemplateManager(testStyles, nil)
		
		// Load templates from embedded filesystem
		err := tm.AddTemplatesFromEmbed(testTemplateFS, "testdata")
		require.NoError(t, err)
		
		// Check that templates were loaded
		templates := tm.ListTemplates()
		assert.Contains(t, templates, "greeting")
		assert.Contains(t, templates, "simple")
		
		// Test rendering a loaded template
		data := struct {
			Name    string
			Message string
		}{
			Name:    "alice",
			Message: "Hello, World!",
		}
		
		expected := testStyles["title"].Render("ALICE") + " says " + testStyles["body"].Render("Hello, World!")
		actual, err := tm.RenderTemplate("greeting", data)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
		
		// Test the simple template
		dateData := struct{ Date string }{Date: "2025-01-01"}
		expectedDate := testStyles["date"].Render("Today is 2025-01-01")
		actualDate, err := tm.RenderTemplate("simple", dateData)
		require.NoError(t, err)
		assert.Equal(t, expectedDate, actualDate)
	})
}

func TestExpandTags(t *testing.T) {
	// a buffer to render to
	var buf bytes.Buffer
	renderer := lipgloss.NewRenderer(&buf)
	lipbalm.SetDefaultRenderer(renderer)

	t.Run("simple styled tag", func(t *testing.T) {
		renderer.SetColorProfile(termenv.TrueColor)
		template := `<title>Hello</title>`
		expected := testStyles["title"].Render("Hello")
		actual, err := lipbalm.ExpandTags(template, testStyles)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("nested tags", func(t *testing.T) {
		renderer.SetColorProfile(termenv.TrueColor)
		template := `<title>Hello <date>World</date></title>`
		expected := testStyles["title"].Render("Hello " + testStyles["date"].Render("World"))
		actual, err := lipbalm.ExpandTags(template, testStyles)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("no-format tag", func(t *testing.T) {
		t.Run("with color", func(t *testing.T) {
			renderer.SetColorProfile(termenv.TrueColor)
			template := `<title>Status</title><no-format> ✓</no-format>`
			expected := testStyles["title"].Render("Status")
			actual, err := lipbalm.ExpandTags(template, testStyles)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})

		t.Run("without color", func(t *testing.T) {
			renderer.SetColorProfile(termenv.Ascii)
			template := `<title>Status</title><no-format> ✓</no-format>`
			expected := "Status ✓"
			actual, err := lipbalm.ExpandTags(template, testStyles)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	})

	t.Run("no lipbalm tags", func(t *testing.T) {
		renderer.SetColorProfile(termenv.TrueColor)
		template := `Just some text.`
		expected := `Just some text.`
		actual, err := lipbalm.ExpandTags(template, testStyles)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("unclosed lipbalm tag", func(t *testing.T) {
		renderer.SetColorProfile(termenv.TrueColor)
		template := `<title>Hello`
		expected := `<title>Hello`
		actual, err := lipbalm.ExpandTags(template, testStyles)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestMain(m *testing.M) {
	// set a dummy renderer for all tests
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(io.Discard))
	m.Run()
}
