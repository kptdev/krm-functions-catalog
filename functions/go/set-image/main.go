package main

import (
	"os"
)

// nolint
func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}
