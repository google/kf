package main

import (
	"fmt"
	"os"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands"
)

func main() {
	if err := commands.NewKfCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
