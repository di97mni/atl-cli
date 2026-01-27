package main

import (
	"os"

	"github.com/martin/atl-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
