// +build !windows

package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/buildpackapplifecycle/shell"
	"code.cloudfoundry.org/buildpackapplifecycle/shell/exec"
	"code.cloudfoundry.org/goshims/osshim"
)

func main() {
	if err := shell.Run(&osshim.OsShim{}, exec.New(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
