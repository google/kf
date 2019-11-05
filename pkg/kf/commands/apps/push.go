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
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/internal/envutil"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/manifest"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/cobra"
	"knative.dev/pkg/ptr"
)

// SrcImageBuilder creates and uploads a container image that contains the
// contents of the argument 'dir'.
type SrcImageBuilder interface {
	BuildSrcImage(dir, srcImage string, filter KontextFilter) error
}

// KontextFilter is used to select which files should be packaged into the
// Kontext container.
type KontextFilter = func(path string) (bool, error)

// SrcImageBuilderFunc converts a func into a SrcImageBuilder.
type SrcImageBuilderFunc func(dir, srcImage string, rebase bool, filter KontextFilter) error

// BuildSrcImage implements SrcImageBuilder.
func (f SrcImageBuilderFunc) BuildSrcImage(dir, srcImage string, filter KontextFilter) error {
	oldPrefix := log.Prefix()
	oldFlags := log.Flags()

	log.SetPrefix("\033[32m[source upload]\033[0m ")
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	log.Printf("Uploading %s to image %s", dir, srcImage)
	err := f(dir, srcImage, false, filter)

	log.SetPrefix(oldPrefix)
	log.SetFlags(oldFlags)
	log.SetOutput(os.Stderr)

	return err
}

// NewPushCommand creates a push command.
func NewPushCommand(
	p *config.KfParams,
	client apps.Client,
	pusher apps.Pusher,
	b SrcImageBuilder,
	serviceBindingClient servicebindings.ClientInterface,
) *cobra.Command {
	var (
		containerRegistry   string
		sourceImage         string
		containerImage      string
		dockerfilePath      string
		manifestFile        string
		instances           int
		minScale            int
		maxScale            int
		path                string
		buildpack           string
		stack               string
		envs                []string
		enableHTTP2         bool
		noManifest          bool
		noStart             bool
		healthCheckType     string
		healthCheckTimeout  int
		startupCommand      string
		containerEntrypoint string
		containerArgs       []string

		// Route Flags
		rawRoutes         []string
		noRoute           bool
		randomRouteDomain bool
	)

	var pushCmd = &cobra.Command{
		Use:   "push APP_NAME",
		Short: "Create a new app or sync changes to an existing app",
		Example: `
  kf push myapp
  kf push myapp --buildpack my.special.buildpack # Discover via kf buildpacks
  kf push myapp --env FOO=bar --env BAZ=foo
  kf push myapp --stack cloudfoundry/cflinuxfs3 # Use a cflinuxfs3 runtime
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

			cmd.SilenceUsage = true

			appName := ""
			if len(args) > 0 {
				appName = args[0]
			}

			var pushManifest *manifest.Manifest
			switch {
			case noManifest:
				if pushManifest, err = manifest.New(appName); err != nil {
					return err
				}
			case manifestFile != "":
				if pushManifest, err = manifest.NewFromFile(manifestFile); err != nil {
					return fmt.Errorf("supplied manifest file %s resulted in error: %v", manifestFile, err)
				}
			default:
				if pushManifest, err = manifest.CheckForManifest(path); err != nil {
					return fmt.Errorf("error checking directory %s for manifest file: %v", path, err)
				}

				if pushManifest == nil {
					if pushManifest, err = manifest.New(appName); err != nil {
						return err
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

			overrides := &manifest.Application{}
			{
				overrides.Docker.Image = containerImage
				overrides.Stack = stack
				overrides.Command = startupCommand
				overrides.Args = containerArgs
				overrides.Entrypoint = containerEntrypoint
				overrides.Dockerfile.Path = dockerfilePath

				// Read environment variables from cli args
				envVars, err := envutil.ParseCLIEnvVars(envs)
				if err != nil {
					return err
				}
				overrides.Env = envutil.EnvVarsToMap(envVars)

				if buildpack != "" {
					overrides.Buildpacks = []string{buildpack}
				}

				overrides.HealthCheckTimeout = healthCheckTimeout

				if healthCheckType != "" {
					overrides.HealthCheckType = healthCheckType
				}

				if len(rawRoutes) > 0 {
					overrides.Routes = nil
					for _, rr := range rawRoutes {
						overrides.Routes = append(overrides.Routes, manifest.Route{
							Route: rr,
						})
					}
				}

				// Only override if the user explicitly set it.
				if cmd.Flags().Lookup("no-route").Changed {
					overrides.NoRoute = &noRoute
				}

				// Only override if the user explicitly set it.
				if cmd.Flags().Lookup("random-route").Changed {
					overrides.RandomRoute = &randomRouteDomain
				}

				// Only override if the user explicitly set it.
				if cmd.Flags().Lookup("instances").Changed {
					overrides.Instances = &instances
				}

				// Only override if the user explicitly set it.
				if cmd.Flags().Lookup("min-scale").Changed {
					overrides.MinScale = &minScale
				}

				// Only override if the user explicitly set it.
				if cmd.Flags().Lookup("max-scale").Changed {
					overrides.MaxScale = &maxScale
				}

				if cmd.Flags().Lookup("enable-http2").Changed {
					overrides.EnableHTTP2 = ptr.Bool(enableHTTP2)
				}

				if cmd.Flags().Lookup("no-start").Changed {
					overrides.NoStart = ptr.Bool(noStart)
				}
			}

			for _, app := range appsToDeploy {
				// Warn the user about unofficial fields they might be using before
				// overriding the manifest.
				if err := app.WarnUnofficialFields(cmd.OutOrStderr()); err != nil {
					return err
				}

				if err := app.Override(overrides); err != nil {
					return err
				}

				if err := app.Validate(context.Background()); err.Error() != "" {
					return err
				}

				defaultDomain, err := spaceDefaultDomain(space)
				if err != nil {
					return err
				}

				routes, err := setupRoutes(space, app)
				if err != nil {
					return err
				}

				container, err := app.ToContainer()
				if err != nil {
					return err
				}

				var randomRouteDomain string
				if app.RandomRoute != nil && *app.RandomRoute {
					randomRouteDomain = defaultDomain
				}

				var defaultRouteDomain string
				if len(routes) == 0 && randomRouteDomain == "" && (app.NoRoute == nil || !*app.NoRoute) {
					defaultRouteDomain = defaultDomain
				}

				pushOpts := []apps.PushOption{
					apps.WithPushNamespace(p.Namespace),
					apps.WithPushRoutes(routes),
					apps.WithPushRandomRouteDomain(randomRouteDomain),
					apps.WithPushDefaultRouteDomain(defaultRouteDomain),
					apps.WithPushAppSpecInstances(app.ToAppSpecInstances()),
					apps.WithPushContainer(container),
				}

				if app.Docker.Image == "" {
					// buildpack or Dockerfile app
					registry := containerRegistry
					switch {
					case registry != "":
						break
					default:
						registry = space.Spec.BuildpackBuild.ContainerRegistry
					}

					var imageName string
					srcPath := filepath.Join(path, app.Path)
					switch {
					case sourceImage != "":
						imageName = sourceImage
					default:
						imageName = apps.JoinRepositoryImage(registry, apps.SourceImageName(p.Namespace, app.Name))

						// Kontext has to have a absolute path.
						srcPath, err = filepath.Abs(srcPath)
						if err != nil {
							return err
						}

						// Sanity check that the Dockerfile is in the source
						if app.Dockerfile.Path != "" {
							absDockerPath := filepath.Join(srcPath, filepath.FromSlash(app.Dockerfile.Path))
							if _, err := os.Stat(absDockerPath); os.IsNotExist(err) {
								fmt.Fprintln(cmd.OutOrStdout(), "app root:", srcPath)
								return fmt.Errorf("the Dockerfile %s couldn't be found under the app root", app.Dockerfile.Path)
							}
						}

						if err := b.BuildSrcImage(srcPath, imageName, buildIgnoreFilter(srcPath)); err != nil {
							return err
						}
					}
					pushOpts = append(pushOpts,
						apps.WithPushSourceImage(imageName),
						apps.WithPushBuildpack(app.Buildpack()),
						apps.WithPushStack(app.Stack),
						apps.WithPushDockerfilePath(app.Dockerfile.Path),
					)
				} else {
					if containerRegistry != "" {
						return errors.New("--container-registry can only be used with source pushes, not containers")
					}
					if app.Buildpack() != "" {
						return errors.New("cannot use buildpack and docker image simultaneously")
					}
					if app.Path != "" {
						return errors.New("cannot use path and docker image simultaneously")
					}

					pushOpts = append(pushOpts, apps.WithPushContainerImage(app.Docker.Image))
				}

				// Bind service if set
				var bindings []v1alpha1.AppSpecServiceBinding
				for _, serviceInstance := range app.Services {
					binding := v1alpha1.AppSpecServiceBinding{
						Instance: serviceInstance,
					}
					bindings = append(bindings, binding)
				}
				pushOpts = append(pushOpts, apps.WithPushServiceBindings(bindings))

				err = pusher.Push(app.Name, pushOpts...)

				cmd.SilenceUsage = !utils.ConfigError(err)

				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	// TODO (#420): Generate flags from manifest

	pushCmd.Flags().StringVar(
		&containerRegistry,
		"container-registry",
		"",
		"Container registry to push sources to. Required for buildpack builds not targeting a Kf space.",
	)

	pushCmd.Flags().StringVarP(
		&path,
		"path",
		"p",
		".",
		"Path to the source code (default: current directory)",
	)

	pushCmd.Flags().StringArrayVarP(
		&envs,
		"env",
		"e",
		nil,
		"Set environment variables. Multiple can be set by using the flag multiple times (e.g., NAME=VALUE).",
	)

	pushCmd.Flags().BoolVar(
		&enableHTTP2,
		"enable-http2",
		false,
		"Setup the container to allow application to use HTTP2 and gRPC.",
	)

	pushCmd.Flags().BoolVar(
		&noManifest,
		"no-manifest",
		false,
		"Ignore the manifest file.",
	)

	pushCmd.Flags().StringVarP(
		&buildpack,
		"buildpack",
		"b",
		"",
		"Skip the 'detect' buildpack step and use the given name.",
	)

	pushCmd.Flags().StringVarP(
		&stack,
		"stack",
		"s",
		"",
		"Base image to use for to use for apps created with a buildpack.",
	)

	pushCmd.Flags().StringVar(
		&sourceImage,
		"source-image",
		"",
		"Kontext image containing the source code.",
	)
	pushCmd.Flags().MarkHidden("source-image")

	pushCmd.Flags().StringVar(
		&containerImage,
		"docker-image",
		"",
		"Docker image to deploy.",
	)

	pushCmd.Flags().StringVar(
		&dockerfilePath,
		"dockerfile",
		"",
		"Path to the Dockerfile to build. Relative to the source root.",
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
		"Number of instances of the app to run (default: 1)",
	)

	pushCmd.Flags().IntVar(
		&minScale,
		"min-scale",
		-1, // -1 represents non-user input
		"Minium number of instances the autoscaler will scale to",
	)

	pushCmd.Flags().IntVar(
		&maxScale,
		"max-scale",
		-1, // -1 represents non-user input
		"Maximum number of instances the autoscaler will scale to",
	)

	pushCmd.Flags().BoolVar(
		&noStart,
		"no-start",
		false,
		"Do not start an app after pushing",
	)

	pushCmd.Flags().StringVarP(
		&healthCheckType,
		"health-check-type",
		"u",
		"",
		"Application health check type (http or port, default: port)",
	)

	pushCmd.Flags().IntVarP(
		&healthCheckTimeout,
		"timeout",
		"t",
		0,
		"Time (in seconds) allowed to elapse between starting up an app and the first healthy response from the app.",
	)

	pushCmd.Flags().BoolVar(
		&noRoute,
		"no-route",
		false,
		"Do not map a route to this app and remove routes from previous pushes of this app",
	)

	pushCmd.Flags().BoolVar(
		&randomRouteDomain,
		"random-route",
		false,
		"Create a random route for this app if the app doesn't have a route.",
	)

	pushCmd.Flags().StringArrayVar(
		&rawRoutes,
		"route",
		nil,
		"Use the routes flag to provide multiple HTTP and TCP routes. Each route for this app is created if it does not already exist.",
	)

	pushCmd.Flags().StringVarP(
		&startupCommand,
		"command",
		"c",
		"",
		"Startup command for the app, this overrides the default command specified by the web process.",
	)

	pushCmd.Flags().StringVar(
		&containerEntrypoint,
		"entrypoint",
		"",
		"Overwrite the default entrypoint of the image. Can't be used with the command flag.",
	)

	pushCmd.Flags().StringArrayVar(
		&containerArgs,
		"args",
		nil,
		"Overwrite the args for the image. Can't be used with the command flag.",
	)

	return pushCmd
}

func createRoute(routeStr, namespace string) (v1alpha1.RouteSpecFields, error) {
	hostname, domain, path, err := parseRouteStr(routeStr)
	if err != nil {
		return v1alpha1.RouteSpecFields{}, err
	}

	return v1alpha1.RouteSpecFields{
		Hostname: hostname,
		Domain:   domain,
		Path:     path,
	}, nil
}

// parseRouteStr parses a route URL into a hostname, domain, and path
func parseRouteStr(routeStr string) (string, string, string, error) {
	u, err := url.Parse(routeStr)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse route: %s", err)
	}
	if u.Scheme == "" {
		// Parsing URLs without schemes causes the hostname and domain to incorrectly be empty.
		// We handle this by assuming the route has a HTTP scheme if scheme is not provided.
		u, err = url.Parse("http://" + routeStr)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to parse route: %s", err)
		}
	}

	parts := strings.SplitN(u.Hostname(), ".", 3)

	var hostname string
	var domain string
	var path string

	if len(parts) == 3 {
		// Has hostname
		if parts[0] == "www" {
			// "www" is a hostname exception.
			// strip hostname and include "www" in the domain
			hostname = ""
			domain = strings.Join(parts, ".")
		} else {
			hostname = parts[0]
			domain = strings.Join(parts[1:], ".")
		}

	} else {
		// Only domain
		hostname = ""
		domain = strings.Join(parts, ".")
	}

	path = u.EscapedPath()

	return hostname, domain, path, nil
}

func spaceDefaultDomain(space *v1alpha1.Space) (string, error) {
	for _, domain := range space.Spec.Execution.Domains {
		if domain.Default {
			return domain.Domain, nil
		}
	}

	return "", errors.New("space does not have a default domain")
}

func setupRoutes(space *v1alpha1.Space, app manifest.Application) (routes []v1alpha1.RouteSpecFields, err error) {
	if app.NoRoute != nil && *app.NoRoute {
		return nil, nil
	}

	for _, route := range app.Routes {
		// Parse route string from URL into hostname, domain, and path
		newRoute, err := createRoute(route.Route, space.Name)
		if err != nil {
			return nil, err
		}
		routes = append(routes, newRoute)
	}

	return routes, nil
}

func buildIgnoreFilter(srcPath string) KontextFilter {
	ignoreFiles := []string{
		".kfignore",
		".cfignore",
	}

	var defaultIgnoreLines = []string{
		".cfignore",
		"/manifest.yml",
		".gitignore",
		".git",
		".hg",
		".svn",
		"_darcs",
		".DS_Store",
	}

	var (
		gitignore *ignore.GitIgnore
		err       error
	)
	for _, ignoreFile := range ignoreFiles {
		gitignore, err = ignore.CompileIgnoreFileAndLines(
			filepath.Join(srcPath, ignoreFile),
			defaultIgnoreLines...,
		)
		if err != nil {
			// Just move on.
			continue
		}

		break
	}

	if gitignore == nil {
		gitignore, err = ignore.CompileIgnoreLines(defaultIgnoreLines...)
		if err != nil {
			return func(string) (bool, error) {
				return false, err
			}
		}
	}

	return func(path string) (bool, error) {
		return !gitignore.MatchesPath(path), nil
	}
}
