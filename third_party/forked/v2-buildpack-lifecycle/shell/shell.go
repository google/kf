// +build !windows

package shell

import (
	"fmt"
	"path/filepath"
	"runtime"

	"code.cloudfoundry.org/buildpackapplifecycle/env"
	"code.cloudfoundry.org/goshims/osshim"
)

type exec interface {
	Exec(dir, launcher, args, command string, environ []string)
}

func Run(os osshim.Os, exec exec, shellArgs []string) error {
	var dir string
	var commands []string

	if len(shellArgs) >= 2 {
		dir = shellArgs[1]
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("Provided app direcory does not exist")
		}
	} else {
		dir = filepath.Join(os.Getenv("HOME"), "app")
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("Could not infer app directory, please provide one")
		}
	}
	if absDir, err := filepath.Abs(dir); err == nil {
		dir = absDir
	}

	if len(shellArgs) >= 3 {
		commands = shellArgs[2:]
	} else {
		commands = []string{"bash"}
	}

	if err := env.CalcEnv(os, dir); err != nil {
		return err
	}

	runtime.GOMAXPROCS(1)

	exec.Exec(dir, launcher, shellArgs[0], commands[0], os.Environ())
	return nil
}

const launcher = `
cd "$1"

if [ -n "$(ls ../profile.d/* 2> /dev/null)" ]; then
  for env_file in ../profile.d/*; do
    source $env_file
  done
fi

if [ -n "$(ls .profile.d/* 2> /dev/null)" ]; then
  for env_file in .profile.d/*; do
    source $env_file
  done
fi

shift

exec bash -c "$@"
`
