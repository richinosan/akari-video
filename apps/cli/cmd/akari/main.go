package main

import (
	"os"

	"github.com/richinosan/akari-video/apps/cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
