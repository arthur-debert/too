package tdh

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/arthur-debert/tdh/pkg/tdh/printer"
)

type Todo struct {
	ID       int64  `json:"id"`
	Text     string `json:"text"`
	Status   string `json:"status"`
	Modified string `json:"modified"`
}

func (t *Todo) MakeOutput(useColor bool) {
	var symbole string
	var color ct.Color

	if t.Status == "done" {
		color = ct.Green
		symbole = printer.OkSign
	} else {
		color = ct.Red
		symbole = printer.KoSign
	}

	hashtagReg := regexp.MustCompile(`#[^\\s]*`)

	spaceCount := 6 - len(strconv.FormatInt(t.ID, 10))

	fmt.Print(strings.Repeat(" ", spaceCount), t.ID, " | ")
	if useColor {
		ct.ChangeColor(color, false, ct.None, false)
	}
	fmt.Print(symbole)
	if useColor {
		ct.ResetColor()
	}
	fmt.Print(" ")
	pos := 0
	for _, token := range hashtagReg.FindAllStringIndex(t.Text, -1) {
		fmt.Print(t.Text[pos:token[0]])
		if useColor {
			ct.ChangeColor(ct.Yellow, false, ct.None, false)
		}
		fmt.Print(t.Text[token[0]:token[1]])
		if useColor {
			ct.ResetColor()
		}
		pos = token[1]
	}
	fmt.Println(t.Text[pos:])
}

// Toggle toggles the todo status between done and pending
func (t *Todo) Toggle() {
	if t.Status == "done" {
		t.Status = "pending"
	} else {
		t.Status = "done"
	}
}
