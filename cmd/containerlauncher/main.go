// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alessio/shellescape"
	"sigs.k8s.io/yaml"
)

const (
	appDir            = "/home/vcap/app"
	stagingInfoFile   = "staging_info.yml"
	metadataIdAddress = "http://metadata.google.internal/computeMetadata/v1/instance/id"
)

func main() {
	logger := log.New(os.Stderr, "launcher: ", log.Lmsgprefix)

	logger.Println("Detecting start command")
	startCommand := getStartCommand(logger)

	log.Println("Initializing environment variables")
	initializeEnv()

	log.Println("Starting application process")
	runProcess(startCommand)
}

func getStartCommand(logger *log.Logger) string {
	var startCommand, startCommandSource string
	switch {
	case len(os.Args) > 1:
		startCommand = shellescape.QuoteCommand(os.Args[1:])
		startCommandSource = "args"
	default:
		stagingCommand, err := stagingInfoFileStartCommand(stagingInfoFile)
		if err != nil {
			log.Fatalf("Invalid staging info: %s\n", err)
		}
		startCommand = stagingCommand
		startCommandSource = stagingInfoFile
	}
	logger.Printf("> Start command from %s %q\n", startCommandSource, startCommand)

	if strings.TrimSpace(startCommand) == "" {
		logger.Fatalln("No start command specified or detected in container")
	}

	return startCommand
}

func stagingInfoFileStartCommand(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	switch {
	case os.IsNotExist(err):
		return "", nil
	case err != nil:
		return "", err
	}

	var info struct {
		StartCommand string `json:"start_command"`
	}

	if err := yaml.Unmarshal(data, &info); err != nil {
		return "", fmt.Errorf("invalid staging info: %w", err)
	}

	return info.StartCommand, nil
}

func initializeEnv() {
	guid := ""

	// Attempt fetching metadata from the metadata server:
	req, err := http.NewRequest(http.MethodGet, metadataIdAddress, nil)
	if err == nil {
		req.Header["Metadata-Flavor"] = []string{"Google"}
		client := http.Client{
			Timeout: 500 * time.Millisecond,
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			id, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				guid = string(id)
			}
		}
	}

	host := "0.0.0.0"
	port := os.Getenv("PORT")
	hostPort := fmt.Sprintf("%s:%s", host, port)

	// Override dynamic environment variables
	for _, env := range []struct {
		Key   string
		Value string
	}{
		{"HOME", appDir},
		{"CF_INSTANCE_ADDR", hostPort},
		{"CF_INSTANCE_GUID", guid},
		{"INSTANCE_GUID", guid},
		{"CF_INSTANCE_INDEX", "0"},
		{"INSTANCE_INDEX", "0"},
		{"CF_INSTANCE_IP", host},
		{"CF_INSTANCE_INTERNAL_IP", host},
		{"VCAP_APP_HOST", host},
		{"CF_INSTANCE_PORT", port},
		{"VCAP_APP_PORT", port},
		{"LANG", "en_US.UTF-8"},
	} {
		os.Setenv(env.Key, env.Value)
	}

	// Set up env for local directories.
	for _, env := range []struct {
		Key  string
		Path string
	}{
		{"TMPDIR", filepath.Join(appDir, "..", "tmp")},
		{"DEPS_DIR", filepath.Join(appDir, "..", "deps")},
	} {
		absDir, err := filepath.Abs(env.Path)
		if err == nil {
			os.Setenv(env.Key, absDir)
		}
	}

	// Augment VCAP_APPLICATION to be consistent with values above.
	{
		vcapApplication := map[string]interface{}{}
		err := json.Unmarshal([]byte(os.Getenv("VCAP_APPLICATION")), &vcapApplication)
		if err == nil {
			vcapApplication["host"] = host
			vcapApplication["instance_id"] = guid
			vcapApplication["instance_index"] = 0

			port, err := strconv.Atoi(port)
			if err == nil {
				vcapApplication["port"] = port
			}

			if augmented, err := json.Marshal(vcapApplication); err == nil {
				os.Setenv("VCAP_APPLICATION", string(augmented))
			}
		}
	}

	// Set environment variables that might be missing
	for _, env := range []struct {
		Key   string
		Value string
	}{
		{"VCAP_SERVICES", "{}"},
		{"DATABASE_URL", ""},
	} {
		if _, found := os.LookupEnv(env.Key); !found {
			os.Setenv(env.Key, env.Value)
		}
	}
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

if [ -f .profile ]; then
  source .profile
fi

shift

exec bash -c "$@"
`

func runProcess(command string) {
	syscall.Exec("/bin/bash", []string{
		"bash",
		"-c",
		launcher,
		os.Args[0],
		appDir,
		command,
	}, os.Environ())
}
