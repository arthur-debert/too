package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// setupHelp configures custom help templates
func setupHelp() {
	// Add template functions
	templateFuncs := template.FuncMap{
		"rpad":                    rpad,
		"bold":                    bold,
		"commandAliases":          commandAliases,
		"trimTrailingWhitespaces": trimTrailingWhitespaces,
	}

	// Set custom help template for root command
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		tmpl, err := template.New("help").Funcs(templateFuncs).Parse(helpTemplate)
		if err != nil {
			fmt.Println(err)
			return
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Print(buf.String())
	})
}

// Template helper functions

func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func bold(s string) string {
	return "\033[1m" + s + "\033[0m"
}

func commandAliases(name string, aliases []string) string {
	if len(aliases) > 0 {
		return fmt.Sprintf("%s,%s", name, strings.Join(aliases, ","))
	}
	return name
}

func trimTrailingWhitespaces(s string) string {
	return strings.TrimRight(s, " \t\n")
}
