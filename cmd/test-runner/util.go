// Copyright 2022 Google LLC
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
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"knative.dev/pkg/logging"
)

// runCommand runs the given command and args. Stderr is left to display to
// the os.Stderr while stdout is captured and returned. If the command fails,
// the stdout will be written to stderr.
func runCommand(
	ctx context.Context,
	cmdName string,
	args []string,
) (stdout io.Reader, wait func() error, err error) {
	logger := logging.FromContext(ctx)

	logger.Infof("Running go %s", strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Env = os.Environ()

	stdout, err = cmd.StdoutPipe()
	if err != nil {
		// XXX: We don't care about the error from reading the buffer.
		data, _ := ioutil.ReadAll(stdout)
		logger.Warn(data)
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		// XXX: We don't care about the error from reading the buffer.
		data, _ := ioutil.ReadAll(stdout)
		logger.Warn(data)
		return nil, nil, err
	}

	return stdout, cmd.Wait, nil
}
