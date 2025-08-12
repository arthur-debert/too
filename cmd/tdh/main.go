package main

import (
	"os"

	"github.com/arthur-debert/tdh/internal/version"

)



func main() {
	if err := Execute(); err != nil {
		os.Exit(1)
	}
} 