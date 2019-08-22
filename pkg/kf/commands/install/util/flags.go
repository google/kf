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
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// InstallCommand is used with SetupFlags. A user *MUST* set RunE (and not
// InstallCommand.Command.RunE).
type InstallCommand struct {
	*cobra.Command
	RunE func(ctx context.Context, cmd *cobra.Command, args []string) error
}

// SetupFlags returns a command with the discovered tags added to it.
func SetupFlags(c InstallCommand) *cobra.Command {
	if c.Command.RunE != nil {
		panic("Don't use cobra.Command's RunE. Instead use the InstallCommand's RunE. We need that context!")
	}

	// Deduce the flags
	ctx := setCaptureFlags(context.Background(), true)
	if err := c.RunE(ctx, c.Command, nil); err != nil {
		panic(fmt.Sprintf("failed to capture flags: %s", err))
	}

	capturedFlags, _ := ctx.Value(captureFlagsType{}).(*pflag.FlagSet)
	c.Command.Flags().AddFlagSet(capturedFlags)

	var (
		verbose bool
		quiet   bool
	)
	c.Command.Flags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"Display any commands ran in the shell",
	)
	c.Command.Flags().BoolVarP(
		&quiet,
		"quiet",
		"q",
		false,
		"Non-interactive mode. This will assume yes to yes-no questions",
	)

	c.Command.RunE = func(cmd *cobra.Command, args []string) error {
		ctx = SetContextOutput(context.Background(), cmd.ErrOrStderr())
		ctx = setVerbosity(ctx, verbose)
		ctx = saveFlagSet(ctx, cmd.Flags())
		ctx = setQuiet(ctx, quiet)

		return c.RunE(ctx, cmd, args)
	}

	return c.Command
}

type captureFlagsType struct{}

// IsCapturingFlags returns true if the command is traversing the installer
// discovring flags.
func IsCapturingFlags(ctx context.Context) bool {
	captureFlags, ok := ctx.Value(captureFlagsType{}).(*pflag.FlagSet)
	return ok && captureFlags != nil
}

func setCaptureFlags(ctx context.Context, captureFlags bool) context.Context {
	return context.WithValue(ctx, captureFlagsType{}, &pflag.FlagSet{})
}

func captureFlag(ctx context.Context, label, typeName string) {
	captureFlags, ok := ctx.Value(captureFlagsType{}).(*pflag.FlagSet)
	if !ok || captureFlags == nil {
		return
	}
	name := flagName(label)
	switch typeName {
	case "string":
		captureFlags.String(name, "", "")
	case "bool":
		captureFlags.Bool(name, false, "")
	default:
		panic("unknown flag type: " + typeName)
	}
}

func captureFlagString(ctx context.Context, label string) {
	captureFlag(ctx, label, "string")
}

func captureFlagBool(ctx context.Context, label string) {
	captureFlag(ctx, label, "bool")
}

type flagSetType struct{}

func saveFlagSet(ctx context.Context, fs *pflag.FlagSet) context.Context {
	return context.WithValue(ctx, flagSetType{}, fs)
}

func getFlag(ctx context.Context, label string) *pflag.Flag {
	flags, ok := ctx.Value(flagSetType{}).(*pflag.FlagSet)
	if !ok || flags == nil {
		return nil
	}
	return flags.Lookup(flagName(label))
}

func getFlagString(ctx context.Context, label string) (string, bool) {
	f := getFlag(ctx, label)
	if f == nil || !f.Changed {
		return "", false
	}

	return f.Value.String(), true
}

func getFlagBool(ctx context.Context, label string) (bool, bool) {
	f := getFlag(ctx, label)
	if f == nil || !f.Changed {
		return false, false
	}

	return f.Value.String() == "true", true
}

func hideFlag(ctx context.Context, name string) {
	captureFlags, ok := ctx.Value(captureFlagsType{}).(*pflag.FlagSet)
	if !ok || captureFlags == nil {
		return
	}
	name = flagName(name)
	captureFlags.MarkHidden(name)
}

var alphaNumeric = regexp.MustCompile(`[^a-zA-Z0-9-]`)
var whiteSpace = regexp.MustCompile(`\s`)

func flagName(label string) string {
	name := whiteSpace.ReplaceAllLiteralString(
		strings.ToLower(label),
		"-",
	)
	name = alphaNumeric.ReplaceAllLiteralString(
		strings.ToLower(name),
		"",
	)
	return name
}
