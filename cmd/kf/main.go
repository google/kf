package main

import (
	"fmt"
	"os"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands"
)

func main() {
	commands.InitializeConfig()

	if err := commands.NewKfCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
