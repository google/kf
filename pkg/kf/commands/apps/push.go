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
	"regexp"
	"strings"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/internal/envutil"
	"github.com/google/kf/pkg/kf/apps"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/commands/utils"
	kfi "github.com/google/kf/pkg/kf/internal/kf"
	"github.com/google/kf/pkg/kf/manifest"
	servicebindings "github.com/google/kf/pkg/kf/service-bindings"
	"github.com/poy/service-catalog/cmd/svcat/output"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
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
func NewPushCommand(
	p *config.KfParams,
	client apps.Client,
	pusher apps.Pusher,
	b SrcImageBuilder,
	serviceBindingClient servicebindings.ClientInterface,
) *cobra.Command {
	var (
		containerRegistry  string
		sourceImage        string
		containerImage     string
		manifestFile       string
		instances          int
		minScale           int
		maxScale           int
		path               string
		buildpack          string
		envs               []string
		grpc               bool
		noManifest         bool
		noStart            bool
		healthCheckType    string
		healthCheckTimeout int
		memoryRequest      *resource.Quantity
		storageRequest     *resource.Quantity
		cpuRequest         *resource.Quantity

		// Route Flags
		rawRoutes         []string
		noRoute           bool
		randomRouteDomain bool
	)

	var pushCmd = &cobra.Command{
		Use:   "push APP_NAME",
		Short: "Push a new app or sync changes to an existing app",
		Example: `
  kf push myapp
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
			}

			for _, app := range appsToDeploy {
				if err := app.Override(overrides); err != nil {
					return err
				}

				exactScale, minScale, maxScale, err := calculateScaleBounds(app.Instances, app.MinScale, app.MaxScale)
				if err != nil {
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

				if app.Memory != "" {
					memStr, err := convertResourceQuantityStr(app.Memory)
					if err != nil {
						return err
					}
					mem, parseErr := resource.ParseQuantity(memStr)
					if parseErr != nil {
						return fmt.Errorf("couldn't parse resource quantity %s: %v", memStr, parseErr)
					}
					memoryRequest = &mem
				}

				if app.DiskQuota != "" {
					storageStr, err := convertResourceQuantityStr(app.DiskQuota)
					if err != nil {
						return err
					}
					storage, parseErr := resource.ParseQuantity(storageStr)
					if parseErr != nil {
						return fmt.Errorf("couldn't parse resource quantity %s: %v", storageStr, parseErr)
					}
					storageRequest = &storage
				}

				if app.CPU != "" {
					cpu, parseErr := resource.ParseQuantity(app.CPU)
					if parseErr != nil {
						return fmt.Errorf("couldn't parse resource quantity %s: %v", app.CPU, parseErr)
					}
					cpuRequest = &cpu
				}

				healthCheck, err := apps.NewHealthCheck(app.HealthCheckType, app.HealthCheckHTTPEndpoint, app.HealthCheckTimeout)
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
					apps.WithPushEnvironmentVariables(app.Env),
					apps.WithPushGrpc(grpc),
					apps.WithPushExactScale(exactScale),
					apps.WithPushMinScale(minScale),
					apps.WithPushMaxScale(maxScale),
					apps.WithPushNoStart(noStart),
					apps.WithPushRoutes(routes),
					apps.WithPushMemory(memoryRequest),
					apps.WithPushDiskQuota(storageRequest),
					apps.WithPushCPU(cpuRequest),
					apps.WithPushHealthCheck(healthCheck),
					apps.WithPushRandomRouteDomain(randomRouteDomain),
					apps.WithPushDefaultRouteDomain(defaultRouteDomain),
				}

				if app.Docker.Image == "" {
					// buildpack app
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
						if err := b.BuildSrcImage(srcPath, imageName); err != nil {
							return err
						}
					}
					pushOpts = append(pushOpts,
						apps.WithPushSourceImage(imageName),
						apps.WithPushBuildpack(app.Buildpack()),
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
				for _, serviceInstance := range app.Services {
					binding, created, err := serviceBindingClient.GetOrCreate(
						serviceInstance,
						app.Name,
						servicebindings.WithCreateBindingName(serviceInstance),
						servicebindings.WithCreateNamespace(p.Namespace))
					if err != nil {
						return err
					}
					if created {
						output.WriteBindingDetails(cmd.OutOrStdout(), binding)
					}

				}

				err = pusher.Push(app.Name, pushOpts...)

				cmd.SilenceUsage = !kfi.ConfigError(err)

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
		"The container registry to push sources to. Required for buildpack builds not targeting a Kf space.",
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

	pushCmd.Flags().IntVar(
		&minScale,
		"min-scale",
		-1, // -1 represents non-user input
		"the minium number of instances the autoscaler will scale to",
	)

	pushCmd.Flags().IntVar(
		&maxScale,
		"max-scale",
		-1, // -1 represents non-user input
		"the maximum number of instances the autoscaler will scale to",
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

	return pushCmd
}

func calculateScaleBounds(instances, minScale, maxScale *int) (exact, min, max *int, err error) {
	switch {
	case instances != nil:
		// Exactly
		if minScale != nil || maxScale != nil {
			return nil, nil, nil, errors.New("couldn't set the -i flag and the minScale/maxScale flags in manifest together")
		}

		return instances, nil, nil, nil
	default:
		// Autoscaling or unset
		return nil, minScale, maxScale, nil
	}
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

// convertResourceQuantityStr converts CF resource quantities into the equivalent k8s quantity strings.
// CF interprets K, M, G, T as binary SI units while k8s interprets them as decimal, so we convert them here
// into the k8s binary SI units (Ki, Mi, Gi, Ti)
func convertResourceQuantityStr(r string) (string, error) {
	// Break down resource quantity string into int and unit of measurement
	// Below method breaks down "50G" into ["50G" "50" "G"]
	parts := cfValidBytesPattern.FindStringSubmatch(strings.TrimSpace(r))
	if len(parts) < 3 {
		return "", errors.New("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB")
	}
	num := parts[1]
	unit := strings.ToUpper(parts[2])
	newUnit := unit
	switch unit {
	case "T":
		newUnit = "Ti"
	case "G":
		newUnit = "Gi"
	case "M":
		newUnit = "Mi"
	case "K":
		newUnit = "Ki"
	}

	return num + newUnit, nil
}

var cfValidBytesPattern = regexp.MustCompile(`(?i)^(-?\d+)([KMGT])B?$`)

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
