// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"bufio"
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/kf/pkg/kf/cli"
)

var (
	// BenchmarkColor is used for the commands to report how long they took.
	BenchmarkColor = color.New(color.FgHiGreen, color.Bold)
)

// Command runs a command with the given context and returns the output's
// lines. If the command fails, then the output is logged. If the context has
// verbose set, then the command is logged before being ran.
func Command(ctx context.Context, name string, args ...string) ([]string, error) {
	if cli.GetVerbosity(ctx) {
		ctx = cli.SetLogPrefix(ctx, name)
		cli.Logf(ctx, "%s %s", name, strings.Join(args, " "))

		start := time.Now()
		defer func() {
			cli.Logf(ctx, BenchmarkColor.Sprintf("%s took %v", name, time.Since(start)))
		}()
	}

	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		cli.Logf(ctx, string(output))
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// Kubectl will run the command and block until its done.
func Kubectl(ctx context.Context, args ...string) ([]string, error) {
	return Command(ctx, "kubectl", args...)
}

// Kf will run the command and block until its done.
func Kf(ctx context.Context, args ...string) ([]string, error) {
	return Command(ctx, "kf", args...)
}

// Git will run the command and block until its done.
func Git(ctx context.Context, args ...string) ([]string, error) {
	return Command(ctx, "git", args...)
}
