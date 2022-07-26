package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"code.cloudfoundry.org/buildpackapplifecycle/env"
	"code.cloudfoundry.org/goshims/osshim"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "%s: received only %d arguments\n", os.Args[0], len(os.Args)-1)
		fmt.Fprintf(os.Stderr, "Usage: %s <app-directory> <start-command> <metadata>", os.Args[0])
		os.Exit(1)
	}

	dir := os.Args[1]
	startCommand := os.Args[2]

	absDir, err := filepath.Abs(dir)
	if err == nil {
		dir = absDir
	}

	var command string
	if startCommand != "" {
		command = startCommand
	} else {
		command, err = startCommandFromStagingInfo("staging_info.yml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid staging info - %s", err)
			os.Exit(1)
		}
	}

	if command == "" {
		fmt.Fprintf(os.Stderr, "%s: no start command specified or detected in droplet", os.Args[0])
		os.Exit(1)
	}

	if err := env.CalcEnv(&osshim.OsShim{}, dir); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(3)
	}

	runtime.GOMAXPROCS(1)
	runProcess(dir, command)
}

type stagingInfo struct {
	StartCommand string `yaml:"start_command"`
}

func startCommandFromStagingInfo(stagingInfoPath string) (string, error) {
	stagingInfoData, err := ioutil.ReadFile(stagingInfoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	info := stagingInfo{}

	err = yaml.Unmarshal(stagingInfoData, &info)
	if err != nil {
		return "", errors.New("invalid YAML")
	}

	return info.StartCommand, nil
}
