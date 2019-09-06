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
	"os"

	"github.com/google/kf/pkg/kf/commands/config"
	build "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/segmentio/textio"
	"k8s.io/client-go/rest"
)

const (
	EmptyNamespaceError = "no space targeted, use 'kf target --space SPACE' to target a space"
)

const (
	AsyncLogSuffix = "\n(This is an asynchronous operation. See https://github.com/google/kf/issues/599 for updates on work to support synchronous commands.)\n"
)

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

// GetBuildConfig returns the build interface.
func GetBuildConfig() build.BuildV1alpha1Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed to get in cluster config: %s", err)
	}
	client, err := build.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to setup build client: %s", err)
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
