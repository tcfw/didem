package main

import (
	"os"

	"github.com/tcfw/didem/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
