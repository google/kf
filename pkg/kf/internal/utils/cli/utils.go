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

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"unicode"

	"github.com/MakeNowJust/heredoc"
	"github.com/fatih/color"
	"github.com/google/kf/v2/pkg/kf/commands/group"
	"github.com/segmentio/textio"
	"github.com/spf13/cobra"
	"knative.dev/pkg/logging"
)

// ConfigErr is used to indicate that the returned error is due to a user's
// invalid configuration.
type ConfigErr struct {
	// Reason holds the error message.
	Reason string
}

// Error implements error.
func (e ConfigErr) Error() string {
	return e.Reason
}

// ConfigError returns true if the error is due to user error.
func ConfigError(err error) bool {
	_, ok := err.(ConfigErr)
	return ok
}

type Config struct {
	Namespace   string
	Args        []string
	AllArgs     []string
	Flags       map[string][]string
	SourceImage string
	Stdout      io.Writer
	Stderr      io.Writer
}

// InBuildParseConfig is used by container images that parse args and flags.
func InBuildParseConfig() Config {
	args, flags := parseCommandLine()
	allArgs := append([]string{}, args...)
	for name, flag := range flags {
		for _, f := range flag {
			allArgs = append(allArgs, fmt.Sprintf("--%s=%q", name, f))
		}
	}

	return Config{
		Namespace:   os.Getenv("NAMESPACE"),
		Flags:       flags,
		Args:        args,
		AllArgs:     allArgs,
		SourceImage: os.Getenv("SOURCE_IMAGE"),
		Stdout:      textio.NewPrefixWriter(os.Stdout, os.Getenv("STDOUT_PREFIX")),
		Stderr:      textio.NewPrefixWriter(os.Stdout, os.Getenv("STDERR_PREFIX")),
	}
}

func parseCommandLine() (args []string, flags map[string][]string) {
	if len(os.Args) != 3 {
		log.Fatalf("invalid number of arguments for container: %#v", os.Args)
	}

	if err := json.Unmarshal([]byte(os.Args[1]), &args); err != nil {
		log.Fatalf("failed to unmarshal args: %s", err)
	}
	if err := json.Unmarshal([]byte(os.Args[2]), &flags); err != nil {
		log.Fatalf("failed to unmarshal flags: %s", err)
	}
	return args, flags
}

// CreateProxy creates a proxy to the specified gateway with the specified host in the request header.
func CreateProxy(w io.Writer, host, gateway string, overrideHeader http.Header) *httputil.ReverseProxy {
	fgBlue := color.New(color.FgBlue)
	logger := log.New(w, fgBlue.Sprintf("[%s via %s] ", host, gateway), log.Ltime)
	director := CreateProxyDirector(host, gateway, overrideHeader)

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			director(req)
			logger.Printf("%s %s\n", req.Method, req.URL.RequestURI())
		},
		ErrorLog: logger,
	}
}

// CreateProxyDirector creates a function that modifies requests being sent to a
// http.RoundTripper so the request is redirected to an IP other than what the
// domain specifies.
func CreateProxyDirector(host, gateway string, overrideHeader http.Header) func(*http.Request) {
	return func(req *http.Request) {
		req.Host = host
		req.URL.Scheme = "http"
		req.URL.Host = gateway

		for k, v := range overrideHeader {
			req.Header[k] = v
		}
	}
}

// PrintCurlExamples lists example HTTP requests the user can send.
func PrintCurlExamples(ctx context.Context, listener net.Listener, host, gateway string) {
	logger := logging.FromContext(ctx)
	logger.Infof("Forwarding requests from %s to %s with host %s", listener.Addr(), gateway, host)
	logger.Info("Example GET:")
	logger.Infof("  curl %s", listener.Addr())
	logger.Info("Example POST:")
	logger.Infof("  curl --request POST %s --data \"POST data\"", listener.Addr())
	logger.Info("Browser link:")
	logger.Infof("  http://%s", listener.Addr())
}

// PrintCurlExamplesNoListener prints CURL examples against the real gateway.
func PrintCurlExamplesNoListener(ctx context.Context, host, gateway string) {
	logger := logging.FromContext(ctx)
	logger.Infof("Requests can be sent to %s with host %s", gateway, host)
	logger.Info("Example GET:")
	logger.Infof("  curl -H \"Host: %s\" http://%s", host, gateway)
	logger.Info("Example POST:")
	logger.Infof("  curl --request POST -H \"Host: %s\" http://%s --data \"POST data\"", host, gateway)
}

// ParseJSONOrFile parses the value as JSON if it's valid or else it tries to
// read the value as a file on the filesystem.
func ParseJSONOrFile(jsonOrFile string) (json.RawMessage, error) {
	if json.Valid([]byte(jsonOrFile)) {
		return AssertJSONMap([]byte(jsonOrFile))
	}

	contents, err := ioutil.ReadFile(jsonOrFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file: %v", err)
	}

	result, err := AssertJSONMap(contents)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse %s as JSON: %v", jsonOrFile, err)
	}

	return result, nil
}

// AssertJSONMap asserts that the string is a JSON map.
func AssertJSONMap(jsonString []byte) (json.RawMessage, error) {
	p := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonString), &p); err != nil {
		return nil, fmt.Errorf("value must be a JSON map, got: %q", jsonString)
	}
	return json.RawMessage(jsonString), nil
}

// SplitTags parses a string of tags into a list by splitting the string by whitespace and commas.
func SplitTags(tagsStr string) []string {
	f := func(c rune) bool {
		return unicode.IsSpace(c) || c == ','
	}
	return strings.FieldsFunc(tagsStr, f)
}

// JoinHeredoc joins documentation that was defined in-line by removing leading
// whitespace and joining blocks of text with two newlines.
func JoinHeredoc(docstrings ...string) string {
	var normalized []string
	for _, doc := range docstrings {
		trimmed := strings.TrimSpace(heredoc.Doc(doc))
		if trimmed == "" {
			continue
		}

		normalized = append(normalized, trimmed)
	}

	return strings.Join(normalized, "\n\n")
}

// PreviewCommandGroup returns a CommandGroup that has a [PREVIEW] prefix. It
// will alter each command to have a PersistentPreRunE that notes that it is
// in preview.
func PreviewCommandGroup(name string, cmds ...*cobra.Command) group.CommandGroup {
	// Wrap the preview commands with a warning message.
	for _, cmd := range cmds {
		addPreviewWarning(cmd, previewWarning)
	}

	return group.CommandGroup{
		Name:     "[PREVIEW] " + name,
		Commands: cmds,
	}
}

const (
	previewWarning      = "This command and feature is in preview and can change in future releases."
	experimentalWarning = "This command and feature is an experiment and can change in future releases."
)

func addPreviewWarning(cmd *cobra.Command, tmpl string) {
	var origRunE func(*cobra.Command, []string) error
	switch {
	case cmd.PersistentPreRunE != nil:
		origRunE = cmd.PersistentPreRunE
	case cmd.PersistentPreRun != nil:
		origRun := cmd.PersistentPreRun
		origRunE = func(innerCommand *cobra.Command, args []string) error {
			origRun(innerCommand, args)
			return nil
		}
		cmd.PersistentPreRun = nil
	default:
		origRunE = func(*cobra.Command, []string) error {
			// NOP
			return nil
		}
	}

	cmd.PersistentPreRunE = func(innerCmd *cobra.Command, args []string) error {
		ctx := innerCmd.Context()
		logging.FromContext(ctx).Warn(tmpl)
		return origRunE(innerCmd, args)
	}
}

// ExperimentalCommandGroup returns a CommandGroup that has a [EXPERIMENTAL]
// prefix. It will alter each command to have a PersistentPreRunE that notes
// that it is an experiment.
func ExperimentalCommandGroup(name string, cmds ...*cobra.Command) group.CommandGroup {
	// Wrap the preview commands with a warning message.
	for _, cmd := range cmds {
		addPreviewWarning(cmd, experimentalWarning)
	}

	return group.CommandGroup{
		Name:     "[EXPERIMENTAL] " + name,
		Commands: cmds,
	}
}
