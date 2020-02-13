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
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

var (
	// PrefixColor should be used when adding color to the prefix of a line
	// (e.g., [some-prefix]).
	PrefixColor = color.New(color.FgHiBlue, color.Bold)
	// LabelColor is used for the prompt and select labels.
	LabelColor = color.New(color.FgHiYellow, color.Bold)
)

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

// SetVerbosity returns a context with the desired verbose setting used with
// Command.
func SetVerbosity(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, verboseType{}, verbose)
}

// GetVerbosity returns the verbose setting from a context.
func GetVerbosity(ctx context.Context) bool {
	v, _ := ctx.Value(verboseType{}).(bool)
	return v
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
