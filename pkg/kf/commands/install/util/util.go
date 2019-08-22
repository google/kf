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
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

var (
	// PrefixColor should be used when adding color to the prefix of a line
	// (e.g., [some-prefix]).
	PrefixColor = color.New(color.FgHiBlue, color.Bold)
	// LabelColor is used for the prompt and select labels.
	LabelColor = color.New(color.FgHiYellow, color.Bold)
	// BenchmarkColor is used for the commands to report how long they took.
	BenchmarkColor = color.New(color.FgHiGreen, color.Bold)
)

// Command runs a command with the given context and returns the output's
// lines. If the command fails, then the output is logged. If the context has
// verbose set, then the command is logged before being ran.
func Command(ctx context.Context, name string, args ...string) ([]string, error) {
	if verbose, ok := ctx.Value(verboseType{}).(bool); ok && verbose {
		ctx = SetLogPrefix(ctx, name)
		Logf(ctx, "%s %s", name, strings.Join(args, " "))

		start := time.Now()
		defer func() {
			Logf(ctx, BenchmarkColor.Sprintf("%s took %v", name, time.Since(start)))
		}()
	}

	if IsCapturingFlags(ctx) {
		return nil, nil
	}

	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		Logf(ctx, string(output))
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

// Searcher implements list.Searcher for promptui.Select. It is case
// insensitive and returns true only if the input string is present.
func Searcher(items []string) func(input string, index int) bool {
	return func(input string, index int) bool {
		item := strings.ToLower(items[index])
		input = strings.TrimSpace(strings.ToLower(input))

		return strings.Contains(item, input)
	}
}

// RandName is a utility function that returns a random name for things like
// spaces or projects.
func RandName(prefix string, args ...interface{}) string {
	return fmt.Sprintf(prefix, args...) + strconv.FormatInt(time.Now().UnixNano(), 36)
}

var (
	// NameRegexp ensures a reasonable name.
	NameRegexp = regexp.MustCompile(`^[a-z][0-9a-zA-Z-]{5,29}$`)
	// HostnameRegexp is from https://stackoverflow.com/a/106223
	HostnameRegexp = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
)

func QuietNotPossibleErr(label string) error {
	return fmt.Errorf("--quiet flag used however '--%s' is required", flagName(label))
}

// NamePrompt asks the user to enter a name. It validates it using NameRegexp.
func NamePrompt(ctx context.Context, label, def string) (string, error) {
	captureFlagString(ctx, label)
	if IsCapturingFlags(ctx) {
		return "", nil
	}

	// See if we got the value from a flag
	if value, ok := getFlagString(ctx, label); ok {
		return value, nil
	}

	if IsQuiet(ctx) {
		return "", QuietNotPossibleErr(label)
	}

	prompt := promptui.Prompt{
		Label: LabelColor.Sprint(label),
		Validate: func(input string) error {
			if !NameRegexp.MatchString(input) {
				return errors.New("invalid name")
			}
			return nil
		},
		Default: def,
	}

	name, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return name, nil
}

// HostnamePrompt asks the user to enter a hostname. It validates it using
// HostnameRegexp.
func HostnamePrompt(ctx context.Context, label, def string) (string, error) {
	captureFlagString(ctx, label)
	if IsCapturingFlags(ctx) {
		return "", nil
	}

	// See if we got the value from a flag
	if value, ok := getFlagString(ctx, label); ok {
		return value, nil
	}

	if IsQuiet(ctx) {
		return "", QuietNotPossibleErr(label)
	}

	prompt := promptui.Prompt{
		Label: LabelColor.Sprint(label),
		Validate: func(input string) error {
			if !HostnameRegexp.MatchString(input) {
				return errors.New("invalid hostname")
			}
			return nil
		},
		Default: def,
	}

	name, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return name, nil
}

// SelectOrCreatePrompt prompts the user to select from the given items or
// create a new entry. It uses Searcher and properly colors the label.
func SelectOrCreatePrompt(
	ctx context.Context,
	label string,
	items ...string,
) (bool, string, error) {
	captureFlagBool(ctx, "create-"+label)

	// See if the user chose to create
	if created, ok := getFlagBool(ctx, "create-"+label); ok && created {
		return created, "", nil
	}

	// See if we got the value from a flag
	if value, ok := getFlagString(ctx, label); ok {
		return false, value, nil
	}

	if IsQuiet(ctx) {
		return false, "", QuietNotPossibleErr(label)
	}

	items = append([]string{"Create New " + label}, items...)

	idx, value, err := SelectPrompt(ctx, label, items...)
	if err != nil {
		return false, "", err
	}

	return idx == 0, value, nil
}

// SelectPrompt prompts the user to select from the given items. It uses
// Searcher and properly colors the label.
func SelectPrompt(
	ctx context.Context,
	label string,
	items ...string,
) (int, string, error) {
	captureFlagString(ctx, label)
	if IsCapturingFlags(ctx) {
		return 0, "", nil
	}

	// See if we got the value from a flag
	if value, ok := getFlagString(ctx, label); ok {
		return -1, value, nil
	}

	if IsQuiet(ctx) {
		return 0, "", QuietNotPossibleErr(label)
	}

	p := promptui.Select{
		Label:             LabelColor.Sprint(label),
		StartInSearchMode: true,
		Searcher:          Searcher(items),
		Items:             items,
	}
	return p.Run()
}

// SelectYesNo promts the user to select between yes and no. It will return
// true if the user selects "yes", and false otherwise.
func SelectYesNo(ctx context.Context, label string) (bool, error) {
	if IsQuiet(ctx) || IsCapturingFlags(ctx) {
		return true, nil
	}

	defer func() {
		// YesNo doesn't get a flag
		hideFlag(ctx, label)
	}()

	_, value, err := SelectPrompt(ctx, label, "yes", "no")
	if err != nil {
		return false, err
	}

	return value == "yes", nil
}

type loggerType struct{}

// SetContextOutput returns a context that holds where Logf should write.
func SetContextOutput(ctx context.Context, out io.Writer) context.Context {
	return context.WithValue(ctx, loggerType{}, out)
}

type loggerPrefixType struct{}

// SetLogPrefix returns a context with the desired prefix that Logf will use.
func SetLogPrefix(ctx context.Context, prefix string) context.Context {
	return context.WithValue(ctx, loggerPrefixType{}, prefix)
}

type verboseType struct{}

// setVerbosity returns a context with the desired verbose setting used with
// Command.
func setVerbosity(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, verboseType{}, verbose)
}

type quietType struct{}

func setQuiet(ctx context.Context, quiet bool) context.Context {
	return context.WithValue(ctx, quietType{}, quiet)
}

// IsQuiet returns true if the --quiet flag was set to true. This implies the
// uesr does not intend to use stdin (like in a CI).
func IsQuiet(ctx context.Context) bool {
	quiet, ok := ctx.Value(quietType{}).(bool)
	return quiet && ok
}

// Logf reads the settings from the given context and logs the given message.
func Logf(ctx context.Context, v string, args ...interface{}) {
	out, ok := ctx.Value(loggerType{}).(io.Writer)
	if !ok {
		return
	}

	if !strings.HasSuffix(v, "\n") {
		v += "\n"
	}

	if prefix, ok := ctx.Value(loggerPrefixType{}).(string); ok {
		v = fmt.Sprintf("[%s] %s", PrefixColor.Sprint(prefix), v)
	}

	fmt.Fprintf(out, v, args...)
}
