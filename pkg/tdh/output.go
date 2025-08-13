package tdh

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/printer"
	ct "github.com/daviddengcn/go-colortext"
)

var hashtagRegex = regexp.MustCompile(`#[^\s]*`)

// MakeOutput formats and prints a single todo item to the console.
// This function handles color formatting and hashtag highlighting.
func MakeOutput(t *models.Todo, useColor bool) {
	var symbol string
	var color ct.Color

	if t.Status == "done" {
		color = ct.Green
		symbol = printer.OkSign
	} else {
		color = ct.Red
		symbol = printer.KoSign
	}

	// Right-align the ID with padding
	spaceCount := 6 - len(strconv.FormatInt(t.ID, 10))
	fmt.Print(strings.Repeat(" ", spaceCount), t.ID, " | ")

	// Print status symbol with color
	if useColor {
		ct.ChangeColor(color, false, ct.None, false)
	}
	fmt.Print(symbol)
	if useColor {
		ct.ResetColor()
	}
	fmt.Print(" ")

	// Print text with hashtag highlighting
	printWithHashtagHighlight(t.Text, useColor)
	fmt.Println()
}

// printWithHashtagHighlight prints text with hashtags highlighted in yellow.
func printWithHashtagHighlight(text string, useColor bool) {
	pos := 0
	for _, match := range hashtagRegex.FindAllStringIndex(text, -1) {
		// Print text before hashtag
		fmt.Print(text[pos:match[0]])

		// Print hashtag with color
		if useColor {
			ct.ChangeColor(ct.Yellow, false, ct.None, false)
		}
		fmt.Print(text[match[0]:match[1]])
		if useColor {
			ct.ResetColor()
		}

		pos = match[1]
	}
	// Print remaining text
	fmt.Print(text[pos:])
}
