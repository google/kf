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

package cli

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
)

type interactiveModeType struct{}

// SetInteractiveMode returns a context that marks it as interactive. This
// allows prompts to make decisions based on if the user wants to interact
// with the CLI interactively or not.
func SetInteractiveMode(ctx context.Context, interactive bool) context.Context {
	return context.WithValue(ctx, interactiveModeType{}, interactive)
}

// GetInteractiveMode returns true if SetInteractiveMode has been set to true
// for the given context.
func GetInteractiveMode(ctx context.Context) bool {
	i, _ := ctx.Value(interactiveModeType{}).(bool)
	return i
}

// interactiveModeDisabledErr is returned by a prompt if InteractiveMode is
// not enabled but a prompt is still requested.
type interactiveModeDisabledErr struct {
	message string
}

// Error implements the error interface.
func (e interactiveModeDisabledErr) Error() string {
	return e.message
}

// IsInteractiveModeErr returns true if the given error is because a prompt
// was requested when interactive mode was disabled.
func IsInteractiveModeErr(err error) bool {
	_, ok := err.(interactiveModeDisabledErr)
	return ok
}

// SelectPrompt prompts the user to select from the given items. It uses
// Searcher and properly colors the label. It fails if the context is not in
// interactive mode.
func SelectPrompt(
	ctx context.Context,
	label string,
	items ...string,
) (int, string, error) {
	if !GetInteractiveMode(ctx) {
		return 0, "", interactiveModeDisabledErr{message: label}
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
// true if the user selects "yes", and false otherwise. It always returns yes
// in non-interactive mode.
func SelectYesNo(ctx context.Context, label string) (bool, error) {
	if !GetInteractiveMode(ctx) {
		return true, nil
	}

	_, value, err := SelectPrompt(ctx, label, "yes", "no")
	if err != nil {
		return false, err
	}

	return value == "yes", nil
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

// NamePrompt asks the user to enter a name. It validates it using NameRegexp.
// It returns the default in non-interactive mode.
func NamePrompt(ctx context.Context, label, def string) (string, error) {
	if !GetInteractiveMode(ctx) {
		return def, nil
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
// HostnameRegexp. It returns the default in non-interactive mode.
func HostnamePrompt(ctx context.Context, label, def string) (string, error) {
	if !GetInteractiveMode(ctx) {
		return def, nil
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
