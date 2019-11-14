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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/google/kf/pkg/kf/commands/config"
	cserving "github.com/google/kf/third_party/knative-serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	serving "github.com/google/kf/third_party/knative-serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/segmentio/textio"
	"k8s.io/client-go/rest"
)

const (
	EmptyNamespaceError = "no space targeted, use 'kf target --space SPACE' to target a space"
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

// KfParams stores everything needed to interact with the user and Knative.
type KfParams struct {
	Output    io.Writer
	Namespace string
}

// GetServingConfig returns the serving interface.
func GetServingConfig() cserving.ServingV1alpha1Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed to get in cluster config: %s", err)
	}
	client, err := serving.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to setup serving client: %s", err)
	}
	return client
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

// ValidateNamespace validate non-empty namespace param
func ValidateNamespace(p *config.KfParams) error {
	if p.Namespace == "" {
		return errors.New(EmptyNamespaceError)
	}
	return nil
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
func CreateProxy(w io.Writer, host, gateway string) *httputil.ReverseProxy {
	// TODO (#698): use color package instead of color code
	logger := log.New(w, fmt.Sprintf("\033[34m[%s via %s]\033[0m ", host, gateway), log.Ltime)

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Host = host
			req.URL.Scheme = "http"
			req.URL.Host = gateway

			logger.Printf("%s %s\n", req.Method, req.URL.RequestURI())
		},
		ErrorLog: logger,
	}
}

// PrintCurlExamples lists example HTTP requests the user can send.
func PrintCurlExamples(w io.Writer, listener net.Listener, host, gateway string) {
	fmt.Fprintf(w, "Forwarding requests from %s to %s with host %s\n", listener.Addr(), gateway, host)
	fmt.Fprintln(w, "Example GET:")
	fmt.Fprintf(w, "  curl %s\n", listener.Addr())
	fmt.Fprintln(w, "Example POST:")
	fmt.Fprintf(w, "  curl --request POST %s --data \"POST data\"\n", listener.Addr())
	fmt.Fprintln(w, "Browser link:")
	fmt.Fprintf(w, "  http://%s\n", listener.Addr())
	fmt.Fprintln(w)
}

// PrintCurlExamplesNoListener prints CURL examples against the real gateway.
func PrintCurlExamplesNoListener(w io.Writer, host, gateway string) {
	fmt.Fprintf(w, "Requests can be sent to %s with host %s\n", gateway, host)
	fmt.Fprintln(w, "Example GET:")
	fmt.Fprintf(w, "  curl -H \"Host: %s\" http://%s\n", host, gateway)
	fmt.Fprintln(w, "Example POST:")
	fmt.Fprintf(w, "  curl --request POST -H \"Host: %s\" http://%s --data \"POST data\"\n", host, gateway)
	fmt.Fprintln(w)
}
