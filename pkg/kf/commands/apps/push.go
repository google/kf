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
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/internal/envutil"
	kfi "github.com/google/kf/pkg/kf/internal/kf"
	"github.com/google/kf/pkg/kf/manifest"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SrcImageBuilder creates and uploads a container image that contains the
// contents of the argument 'dir'.
type SrcImageBuilder interface {
	BuildSrcImage(dir, srcImage string) error
}

// SrcImageBuilderFunc converts a func into a SrcImageBuilder.
type SrcImageBuilderFunc func(dir, srcImage string, rebase bool) error

// BuildSrcImage implements SrcImageBuilder.
func (f SrcImageBuilderFunc) BuildSrcImage(dir, srcImage string) error {
	oldPrefix := log.Prefix()
	oldFlags := log.Flags()

	log.SetPrefix("\033[32m[source upload]\033[0m ")
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	log.Printf("Uploading %s to image %s", dir, srcImage)
	err := f(dir, srcImage, false)

	log.SetPrefix(oldPrefix)
	log.SetFlags(oldFlags)
	log.SetOutput(os.Stderr)

	return err
}

// NewPushCommand creates a push command.
func NewPushCommand(p *config.KfParams, client apps.Client, pusher apps.Pusher, b SrcImageBuilder) *cobra.Command {
	var (
		containerRegistry string
		sourceImage       string
		containerImage    string
		manifestFile      string
		instances         int
		serviceAccount    string
		path              string
		buildpack         string
		envs              []string
		grpc              bool
		routes            []v1alpha1.Route
	)

	var pushCmd = &cobra.Command{
		Use:   "push APP_NAME",
		Short: "Push a new app or sync changes to an existing app",
		Example: `
  kf push myapp
  kf push myapp --container-registry gcr.io/myproject
  kf push myapp --buildpack my.special.buildpack # Discover via kf buildpacks
  kf push myapp --env FOO=bar --env BAZ=foo
  `,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := utils.ValidateNamespace(p); err != nil {
				return err
			}

			space, err := p.GetTargetSpaceOrDefault()
			if err != nil {
				return err
			}

			if containerImage != "" {
				if containerRegistry != "" {
					return errors.New("cannot use --container-registry and --docker-image simultaneously")
				}
				if buildpack != "" {
					return errors.New("cannot use --buildpack and --docker-image simultaneously")
				}
				if path != "." { // the default value
					return errors.New("cannot use --path and --docker-image simultaneously")
				}
			}

			switch {
			case containerRegistry != "":
				break
			case space.Spec.BuildpackBuild.ContainerRegistry != "":
				containerRegistry = space.Spec.BuildpackBuild.ContainerRegistry
			default:
				if containerImage == "" {
					return errors.New("container-registry is required for buildpack apps")
				}
			}

			cmd.SilenceUsage = true

			appName := ""
			if len(args) > 0 {
				appName = args[0]
			}

			// Kontext has to have a absolute path.
			path, err = filepath.Abs(path)
			if err != nil {
				return err
			}

			var pushManifest *manifest.Manifest
			if manifestFile != "" {
				if pushManifest, err = manifest.NewFromFile(manifestFile); err != nil {
					return fmt.Errorf("supplied manifest file %s resulted in error: %v", manifestFile, err)
				}
			} else {
				pushManifest, err = manifest.CheckForManifest(path)
				if err != nil {
					return fmt.Errorf("error checking directory %s for manifest file: %v", path, err)
				}

				if pushManifest == nil {
					// Use a default manifest
					if pushManifest, err = manifest.New(appName); err != nil {
						return errors.New("an app name is required if there is no manifest file")
					}
				}
			}

			appsToDeploy := pushManifest.Applications
			if appName != "" {
				// deploy one app from the manifest
				app, err := pushManifest.App(appName)
				if err != nil {
					return err
				}

				appsToDeploy = []manifest.Application{*app}
			}

			var minScale int
			var maxScale int
			for _, app := range appsToDeploy {
				minScale, maxScale, err = calculateScaleBounds(instances, app.MinScale, app.MaxScale)
				if err != nil {
					return err
				}

				// Read environment variables from cli args
				envVars, err := envutil.ParseCLIEnvVars(envs)
				if err != nil {
					return err
				}
				envMap := envutil.EnvVarsToMap(envVars)

				var imageName string
				if containerImage == "" {

					srcPath := filepath.Join(path, app.Path)
					switch {
					case sourceImage != "":
						imageName = sourceImage
					default:
						imageName = apps.JoinRepositoryImage(containerRegistry, apps.SourceImageName(p.Namespace, app.Name))

						if err := b.BuildSrcImage(srcPath, imageName); err != nil {
							return err
						}
					}

					if app.Env == nil {
						app.Env = make(map[string]string)
					}

					// Merge cli arg environment variables over manifest ones
					for k, v := range envMap {
						app.Env[k] = v
					}

					envMap = app.Env
				}

				manifestRoutes := app.Routes
				for _, route := range manifestRoutes {
					// Parse route string from URL into hostname, domain, and path
					routes = append(routes, createRoute(route.Route, p.Namespace))
				}

				err = pusher.Push(app.Name,
					apps.WithPushSourceImage(imageName),
					apps.WithPushContainerImage(containerImage),
					apps.WithPushNamespace(p.Namespace),
					apps.WithPushContainerRegistry(containerRegistry),
					apps.WithPushServiceAccount(serviceAccount),
					apps.WithPushEnvironmentVariables(envMap),
					apps.WithPushGrpc(grpc),
					apps.WithPushBuildpack(buildpack),
					apps.WithPushMinScale(minScale),
					apps.WithPushMaxScale(maxScale),
					apps.WithPushRoutes(routes),
				)

				cmd.SilenceUsage = !kfi.ConfigError(err)

				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	pushCmd.Flags().StringVar(
		&containerRegistry,
		"container-registry",
		"",
		"The container registry to push containers. Required if not targeting a Kf space.",
	)

	pushCmd.Flags().StringVar(
		&serviceAccount,
		"service-account",
		"",
		"The service account to enable access to the container registry",
	)

	pushCmd.Flags().StringVarP(
		&path,
		"path",
		"p",
		".",
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

	pushCmd.Flags().StringVar(
		&sourceImage,
		"source-image",
		"",
		"The kontext image that has the source code.",
	)
	pushCmd.Flags().MarkHidden("source-image")

	pushCmd.Flags().StringVar(
		&containerImage,
		"docker-image",
		"",
		"The docker image to deploy.",
	)

	pushCmd.Flags().StringVarP(
		&manifestFile,
		"manifest",
		"f",
		"",
		"Path to manifest",
	)

	pushCmd.Flags().IntVarP(
		&instances,
		"instances",
		"i",
		-1, // -1 represents non-user input
		"the number of instances (default is 1)",
	)

	return pushCmd
}

func calculateScaleBounds(instances int, minScale, maxScale *int) (int, int, error) {
	zero := 0
	if instances != -1 {
		if minScale != nil || maxScale != nil {
			return -1, -1, errors.New("couldn't set the -i flag and the minScale/maxScale flags in manifest together")
		}
		return instances, instances, nil
	} else {
		if minScale == nil && maxScale == nil {
			// both default bounds are 1
			return 1, 1, nil
		}

		// Set 0 as default value(unbound) if one of min or max is not set
		if minScale == nil {
			minScale = &zero
		}

		if maxScale == nil {
			maxScale = &zero
		}

		return *minScale, *maxScale, nil
	}

}

func createRoute(appName string, routeStr string, ns string) v1alpha1.Route {
	u, err := url.Parse(routeStr)
	if err != nil {
		panic(err)
	}

	parts := strings.SplitN(u.Hostname(), ".", 3)

	var hostname string
	var domain string
	var urlPath string

	if len(parts) == 3 {
		// Has hostname
		hostname = parts[0]
		domain = strings.Join(parts[1:], ".")
	} else {
		// Only domain
		hostname = ""
		domain = strings.Join(parts, ".")
	}

	urlPath = u.EscapedPath()

	r := &v1alpha1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind: "Route",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name: v1alpha1.GenerateName(
				hostname,
				domain,
				urlPath,
			),
		},
		Spec: v1alpha1.RouteSpec{
			Hostname:            hostname,
			Domain:              domain,
			Path:                urlPath,
			KnativeServiceNames: []string{appName},
		},
	}

	return r
}
