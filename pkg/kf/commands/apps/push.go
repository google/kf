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

package apps

import (
	"path/filepath"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	kfi "github.com/GoogleCloudPlatform/kf/pkg/kf/internal/kf"
	"github.com/spf13/cobra"
)

// NewPushCommand creates a push command.
func NewPushCommand(p *config.KfParams, pusher kf.Pusher) *cobra.Command {
	var (
		containerRegistry string
		dockerImage       string
		serviceAccount    string
		path              string
		buildpack         string
		envs              []string
		grpc              bool
	)

	var pushCmd = &cobra.Command{
		Use:   "push APP_NAME [--container-registry CONTAINER_REGISTRY]",
		Short: "Push a new app or sync changes to an existing app",
		Example: `
  kf push myapp --container-registry gcr.io/myproject
  kf push myapp --container-registry docker.io/myuser --service-account docker-account
  kf push myapp --container-registry gcr.io/myproject --buildpack my.special.buildpack # Discover via kf buildpacks
  kf push myapp --container-registry gcr.io/myproject --env FOO=bar --env BAZ=foo
  kf push myapp --docker-image docker.io/myuser/myimage
  `,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Cobra ensures we are only called with a single argument.
			appName := args[0]

			if path != "" {
				var err error
				path, err = filepath.Abs(path)
				if err != nil {
					return err
				}
			}

			err := pusher.Push(
				appName,
				kf.WithPushNamespace(p.Namespace),
				kf.WithPushContainerRegistry(containerRegistry),
				kf.WithPushDockerImage(dockerImage),
				kf.WithPushServiceAccount(serviceAccount),
				kf.WithPushPath(path),
				kf.WithPushEnvironmentVariables(envs),
				kf.WithPushGrpc(grpc),
				kf.WithPushBuildpack(buildpack),
			)
			cmd.SilenceUsage = !kfi.ConfigError(err)

			return err
		},
	}

	pushCmd.Flags().StringVar(
		&containerRegistry,
		"container-registry",
		"",
		"The container registry to push containers. Either docker-image or container-registry must be set (but not both).",
	)

	pushCmd.Flags().StringVar(
		&dockerImage,
		"docker-image",
		"",
		"The docker image to push. Either docker-image or container-registry must be set (but not both).",
	)

	pushCmd.Flags().StringVar(
		&serviceAccount,
		"service-account",
		"",
		"The service account to enable access to the container registry",
	)

	pushCmd.Flags().StringVar(
		&path,
		"path",
		"",
		"The path the source code lives. Defaults to current directory.",
	)

	pushCmd.Flags().StringArrayVarP(
		&envs,
		"env",
		"e",
		nil,
		"Set environment variables. Multiple can be set by using the flag multiple times (e.g., NAME=VALUE).",
	)

	pushCmd.Flags().BoolVar(
		&grpc,
		"grpc",
		false,
		"Setup the container to allow application to use gRPC.",
	)

	pushCmd.Flags().StringVarP(
		&buildpack,
		"buildpack",
		"b",
		"",
		"Skip the 'detect' buildpack step and use the given name.",
	)

	return pushCmd
}
