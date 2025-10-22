package main

import (
	"os"

	"github.com/alejandro/technical_test_uvigo/cmd/iot-cli/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}

