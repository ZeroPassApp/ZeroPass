// Package main is the entry point for the ZeroPass CLI.
package main

import (
	"os"

	"github.com/zeropass/zeropass/packages/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
