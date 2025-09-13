package lipbalm

import (
	"strings"
	"text/template"
)

// DefaultTemplateFuncs returns common template functions useful for CLI applications
// These supplement the Sprig functions that are already included
func DefaultTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// String manipulation beyond Sprig
		"indent": func(level int) string {
			return strings.Repeat("  ", level)
		},
		"lines": func(text string) []string {
			return strings.Split(text, "\n")
		},
		"wrapText": func(text string, width int) string {
			// Simple word wrapping
			if width <= 0 || len(text) <= width {
				return text
			}
			
			var result strings.Builder
			words := strings.Fields(text)
			lineLen := 0
			
			for i, word := range words {
				wordLen := len(word)
				if i > 0 && lineLen+wordLen+1 > width {
					result.WriteString("\n")
					lineLen = 0
				} else if i > 0 {
					result.WriteString(" ")
					lineLen++
				}
				result.WriteString(word)
				lineLen += wordLen
			}
			
			return result.String()
		},
		
		// Common formatting helpers
		"pluralize": func(count int, singular, plural string) string {
			if count == 1 {
				return singular
			}
			return plural
		},
		"yesno": func(value bool, yes, no string) string {
			if value {
				return yes
			}
			return no
		},
		
		// Message level helpers
		"messageLevel": func(count int) string {
			if count == 0 {
				return "warning"
			}
			return "info"
		},
	}
}

// MergeTemplateFuncs combines multiple template function maps
func MergeTemplateFuncs(funcMaps ...template.FuncMap) template.FuncMap {
	result := make(template.FuncMap)
	for _, m := range funcMaps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}