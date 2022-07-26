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
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/envutil"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/manifest"
	"github.com/google/kf/v2/pkg/sourceimage"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/ptr"
	"sigs.k8s.io/yaml"
)

// SrcImageBuilder creates and uploads a container image that contains the
// contents of the argument 'dir'.
type SrcImageBuilder interface {
	BuildSrcImage(dir, srcImage string, filter sourceimage.FileFilter) (string, error)
}

type pushParams struct {
	containerRegistry       string
	sourceImage             string
	containerImage          string
	dockerfilePath          string
	manifestFile            string
	instances               int32
	path                    string
	buildpack               string
	stack                   string
	envs                    []string
	noManifest              bool
	noStart                 bool
	healthCheckType         string
	healthCheckTimeout      int
	healthCheckHTTPEndpoint string
	startupCommand          string
	containerEntrypoint     string
	containerArgs           []string
	diskQuota               string
	memoryLimit             string
	cpu                     string

	// Route Flags
	rawRoutes         []string
	noRoute           bool
	randomRouteDomain bool

	// Task Flags
	task bool

	// Variable subtitution
	vars      map[string]string
	varsFiles []string

	appSuffix string
}

// DefaultSrcImageBuilder is the default image builder that implements
// SrcImageBuilderFunc.
func DefaultSrcImageBuilder(dir, srcImage string, filter sourceimage.FileFilter) (string, error) {
	image, err := sourceimage.PackageSourceDirectory(dir, filter)
	if err != nil {
		return "", err
	}

	ref, err := sourceimage.PushImage(srcImage, image, false)
	if err != nil {
		return "", err
	}

	return ref.String(), nil
}

// SrcImageBuilderFunc converts a func into a SrcImageBuilder.
type SrcImageBuilderFunc func(dir, srcImage string, filter sourceimage.FileFilter) (string, error)

// BuildSrcImage implements SrcImageBuilder.
func (f SrcImageBuilderFunc) BuildSrcImage(dir, srcImage string, filter sourceimage.FileFilter) (string, error) {
	return f(dir, srcImage, filter)
}

// NewPushCommand creates a push command.
func NewPushCommand(
	p *config.KfParams,
	pusher apps.Pusher,
	b SrcImageBuilder,
) *cobra.Command {
	var params pushParams

	var pushCmd = &cobra.Command{
		Use:   "push APP_NAME",
		Short: "Create a new App or apply updates to an existing one.",
		Example: `
  kf push myapp
  kf push myapp --buildpack my.special.buildpack # Discover via kf buildpacks
  kf push myapp --env FOO=bar --env BAZ=foo
  kf push myapp --stack cloudfoundry/cflinuxfs3 # Use a cflinuxfs3 runtime
  kf push myapp --health-check-http-endpoint /myhealthcheck # Specify a healthCheck for the app
  `,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()

			timeout, err := pushTimeout(os.LookupEnv)
			if err != nil {
				return fmt.Errorf("failed to parse timeout env: %v", err)
			}

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			defer func() {
				cmd.SilenceUsage = !utils.ConfigError(err)
			}()

			if err := p.ValidateSpaceTargeted(); err != nil {
				return err
			}

			space, err := p.GetTargetSpace(ctx)
			if err != nil {
				return err
			}

			appName := ""
			if len(args) > 0 {
				appName = args[0]
			}

			appDevExBuilds := p.FeatureFlags(ctx).AppDevExperienceBuilds().IsEnabled()

			pushManifest, err := createManifest(ctx, &params, appName)
			if err != nil {
				return err
			}

			appsToDeploy, err := createAppsToDeploy(*pushManifest, appName)
			if err != nil {
				return err
			}

			overrides, err := createAppOverrides(cmd, &params)
			if err != nil {
				return err
			}

			for _, app := range appsToDeploy {
				// Warn the user about unofficial fields they might be using before
				// overriding the manifest.
				if err := app.WriteWarnings(ctx); err != nil {
					return err
				}

				if err := app.Override(overrides); err != nil {
					return err
				}

				if err := app.Validate(ctx); err.Error() != "" {
					return err
				}

				stack, err := checkStackExists(ctx, p, app, space)
				if err != nil {
					return err
				}

				routes, err := setupRoutes(space, app)
				if err != nil {
					return err
				}

				generateRandomRoute := false
				if app.RandomRoute != nil && *app.RandomRoute {
					generateRandomRoute = true
				}

				generateDefaultRoute := false
				if len(routes) == 0 &&
					!generateRandomRoute &&
					(app.NoRoute == nil || !*app.NoRoute) &&
					(app.Task == nil || !*app.Task) {
					generateDefaultRoute = true
				}

				var image *string

				if params.containerImage != "" {
					image = &params.containerImage
				} else if app.Docker.Image != "" {
					image = &app.Docker.Image
				}

				if image != nil {
					if params.containerRegistry != "" {
						return errors.New("--container-registry can only be used with source pushes, not containers")
					}
					if params.buildpack != "" {
						return errors.New("cannot use buildpack and docker image simultaneously")
					}
					if app.Path != "" {
						return errors.New("cannot use path and docker image simultaneously")
					}
				}

				if params.containerRegistry != "" && appDevExBuilds {
					return errors.New("--container-registry is not valid with AppDevExperienceBuilds")
				}

				container, err := app.ToContainer()
				if err != nil {
					return err
				}

				pushOpts := []apps.PushOption{
					apps.WithPushSpace(p.Space),
					apps.WithPushRoutes(routes),
					apps.WithPushGenerateRandomRoute(generateRandomRoute),
					apps.WithPushGenerateDefaultRoute(generateDefaultRoute),
					apps.WithPushAppSpecInstances(app.ToAppSpecInstances()),
					apps.WithPushContainer(container),
					apps.WithPushContainerImage(image),
				}

				var srcPath string
				if params.path != "" {
					srcPath = params.path
				} else {
					srcPath = filepath.Join(pushManifest.RelativePathRoot, app.Path)
				}

				if !appDevExBuilds {
					registry := params.sourceRegistryName(*space)

					var sourceImage string
					if params.sourceImage != "" {
						sourceImage = params.sourceImage
					} else {
						sourceImage = apps.JoinRepositoryImage(registry, apps.SourceImageName(p.Space, app.Name))
					}

					builder, shouldPushSource, err := app.DetectBuildType(space.Status.BuildConfig)
					if err != nil {
						return err
					}

					if !shouldPushSource && params.containerRegistry != "" {
						return errors.New("--container-registry can only be used with source pushes, not containers")
					}

					legacyPush := params.containerRegistry != "" || params.sourceImage != ""

					if shouldPushSource && legacyPush {
						// Legacy source upload path.
						// TODO: This is still here to ensure we haven't
						// broken the existing CLI UX. When we move to v3.x.x
						// we can remove all this.

						imageWithDigest, err := pushSourceImage(ctx, cmd, b, app, srcPath, sourceImage)
						if err != nil {
							return err
						}
						sourceImage = imageWithDigest
					}

					buildSpec, err := builder(sourceImage)
					if err != nil {
						return err
					}

					if shouldPushSource && !legacyPush {
						// Normal source upload.

						// We don't need the builder's source image. The
						// controller doesn't expect both a source image and a
						// SourcePackage. Therefore we need to remove it.
						for i, p := range buildSpec.Params {
							if p.Name != v1alpha1.SourceImageParamName {
								continue
							}

							buildSpec.Params = append(buildSpec.Params[:i], buildSpec.Params[i+1:]...)
							break
						}

						// Add the source path option.
						pushOpts = append(pushOpts, apps.WithPushSourcePath(srcPath))
					}

					// Add the build spec.
					pushOpts = append(pushOpts, apps.WithPushBuild(buildSpec))
				} else if params.containerImage == "" && app.Docker.Image == "" {
					// AppDevExperience builds.
					// The only time we wouldn't want the user pushing source
					// code is if they used the `--docker-image` flag or if
					// the App manifest sets a docker image. Push source code.
					//
					// NOTE: Once/if we decide decide that AppDevExperience is
					// thee ONLY way for customers to build source into
					// containers, then we can delete the other branch.
					logging.FromContext(ctx).Warn("Using AppDevExperience Builds. This is a preview feature and is subject to change.")

					pushOpts = append(pushOpts,
						apps.WithPushADXBuild(true),
						apps.WithPushSourcePath(srcPath),
						apps.WithPushADXStack(stack),
						apps.WithPushADXContainerRegistry(space.Status.BuildConfig.ContainerRegistry),
					)

					if params.dockerfilePath != "" {
						pushOpts = append(pushOpts,
							apps.WithPushADXDockerfile(params.dockerfilePath),
						)
					}
				}

				// Bind services if set
				var bindings []v1alpha1.ServiceInstanceBinding
				for _, serviceInstance := range app.Services {
					binding := createServiceInstanceBinding(app.Name, serviceInstance, p.Space)
					bindings = append(bindings, binding)
				}
				pushOpts = append(pushOpts, apps.WithPushServiceBindings(bindings))

				err = pusher.Push(ctx, app.Name+params.appSuffix, pushOpts...)

				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	// TODO (#420): Generate flags from manifest

	pushCmd.Flags().StringVar(
		&params.containerRegistry,
		"container-registry",
		"",
		"Container registry to push images to.",
	)

	pushCmd.Flags().StringVarP(
		&params.path,
		"path",
		"p",
		"",
		"If specified, overrides the path to the source code.",
	)

	pushCmd.Flags().StringArrayVarP(
		&params.envs,
		"env",
		"e",
		nil,
		"Set environment variables. Multiple can be set by using the flag multiple times (for example, NAME=VALUE).",
	)

	pushCmd.Flags().BoolVar(
		&params.noManifest,
		"no-manifest",
		false,
		"Do not read the manifest file even if one exists.",
	)

	pushCmd.Flags().StringVarP(
		&params.buildpack,
		"buildpack",
		"b",
		"",
		"Use the specified buildpack rather than the built-in.",
	)

	pushCmd.Flags().StringVarP(
		&params.stack,
		"stack",
		"s",
		"",
		"Base image to use for to use for Apps created with a buildpack.",
	)

	pushCmd.Flags().StringVar(
		&params.sourceImage,
		"source-image",
		"",
		"Image containing the source code.",
	)
	pushCmd.Flags().MarkHidden("source-image")

	pushCmd.Flags().StringVar(
		&params.containerImage,
		"docker-image",
		"",
		"Docker image to deploy rather than building from source.",
	)

	pushCmd.Flags().StringVar(
		&params.dockerfilePath,
		"dockerfile",
		"",
		"Path to the Dockerfile to build. Relative to the source root.",
	)

	pushCmd.Flags().StringVarP(
		&params.manifestFile,
		"manifest",
		"f",
		"",
		"Path to the application manifest.",
	)

	pushCmd.Flags().Int32VarP(
		&params.instances,
		"instances",
		"i",
		-1, // -1 represents non-user input
		"If set, overrides the number of instances of the App to run, -1 represents non-user input.",
	)

	pushCmd.Flags().BoolVar(
		&params.noStart,
		"no-start",
		false,
		"Build but do not run the App.",
	)

	pushCmd.Flags().StringVarP(
		&params.healthCheckType,
		"health-check-type",
		"u",
		"",
		"App health check type: http, port (default) or process.",
	)

	pushCmd.Flags().IntVarP(
		&params.healthCheckTimeout,
		"timeout",
		"t",
		0,
		"Amount of time the App can be unhealthy before declaring it as unhealthy.",
	)

	pushCmd.Flags().StringVar(
		&params.healthCheckHTTPEndpoint,
		"health-check-http-endpoint",
		"",
		"HTTP endpoint to target as part of the health-check. Only valid if health-check-type is http.",
	)

	pushCmd.Flags().BoolVar(
		&params.noRoute,
		"no-route",
		false,
		"Prevents the App from being reachable once deployed.",
	)

	pushCmd.Flags().BoolVar(
		&params.randomRouteDomain,
		"random-route",
		false,
		"Create a random Route for this App if it doesn't have one.",
	)

	pushCmd.Flags().StringArrayVar(
		&params.rawRoutes,
		"route",
		nil,
		"Use the routes flag to provide multiple HTTP and TCP routes. Each Route for this App is created if it does not already exist.",
	)

	pushCmd.Flags().StringVarP(
		&params.startupCommand,
		"command",
		"c",
		"",
		"Startup command for the App, this overrides the default command specified by the web process.",
	)

	pushCmd.Flags().StringVar(
		&params.containerEntrypoint,
		"entrypoint",
		"",
		"Overwrite the default entrypoint of the image. Can't be used with the command flag.",
	)

	pushCmd.Flags().StringArrayVar(
		&params.containerArgs,
		"args",
		nil,
		"Override the args for the image. Can't be used with the command flag.",
	)

	pushCmd.Flags().StringVarP(
		&params.diskQuota,
		"disk-quota",
		"k",
		"",
		"Size of dedicated ephemeral disk attached to each App instance (for example 512M, 2G, 1T).",
	)

	pushCmd.Flags().StringVarP(
		&params.memoryLimit,
		"memory-limit",
		"m",
		"",
		"Amount of dedicated RAM to give each App instance (for example 512M, 6G, 1T).",
	)

	pushCmd.Flags().StringVar(
		&params.cpu,
		"cpu-cores",
		"",
		"Number of dedicated CPU cores to give each App instance (for example 100m, 0.5, 1, 2). For more information see https://kubernetes.io/docs/tasks/configure-pod-container/assign-cpu-resource/.",
	)

	pushCmd.Flags().StringToStringVar(
		&params.vars,
		"var",
		nil,
		"Manifest variable substitution. Multiple can be set by using the flag multiple times (for example NAME=VALUE).",
	)
	pushCmd.Flags().StringArrayVar(
		&params.varsFiles,
		"vars-file",
		nil,
		"JSON or YAML file to read variable substitutions from. Can be supplied multiple times.",
	)

	pushCmd.Flags().StringVar(
		&params.appSuffix,
		"app-suffix",
		"",
		"Suffix to append to the end of every pushed App.",
	)

	pushCmd.Flags().BoolVar(
		&params.task,
		"task",
		false,
		"Push an App to execute Tasks only. The App will be built, but not run. It will not have a route assigned.",
	)

	return pushCmd
}

func pushTimeout(lookupEnv func(string) (string, bool)) (time.Duration, error) {
	for _, env := range []string{"CF_STARTUP_TIMEOUT", "KF_STARTUP_TIMEOUT"} {
		value, ok := lookupEnv(env)
		if !ok {
			continue
		}

		return time.ParseDuration(value)
	}

	return 15 * time.Minute, nil
}

func createServiceInstanceBinding(appName, serviceInstance, namespace string) v1alpha1.ServiceInstanceBinding {
	return v1alpha1.ServiceInstanceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v1alpha1.MakeServiceBindingName(appName, serviceInstance),
			Namespace: namespace,
		},
		Spec: v1alpha1.ServiceInstanceBindingSpec{
			BindingType: v1alpha1.BindingType{
				App: &v1alpha1.AppRef{
					Name: appName,
				},
			},
			InstanceRef: v1.LocalObjectReference{
				Name: serviceInstance,
			},
			ParametersFrom: v1.LocalObjectReference{
				Name: v1alpha1.MakeServiceBindingParamsSecretName(appName, serviceInstance),
			},
		},
	}
}

func createRoute(route manifest.Route, namespace string) (v1alpha1.RouteWeightBinding, error) {
	hostname, domain, path, err := parseRouteStr(route.Route)
	if err != nil {
		return v1alpha1.RouteWeightBinding{}, err
	}

	rwb := v1alpha1.RouteWeightBinding{
		RouteSpecFields: v1alpha1.RouteSpecFields{
			Hostname: hostname,
			Domain:   domain,
			Path:     path,
		},
	}

	if route.AppPort != 0 {
		rwb.DestinationPort = ptr.Int32(route.AppPort)
	}

	return rwb, nil
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

func setupRoutes(space *v1alpha1.Space, app manifest.Application) (routes []v1alpha1.RouteWeightBinding, err error) {
	if app.Task != nil && *app.Task {
		return nil, nil
	}

	if app.NoRoute != nil && *app.NoRoute {
		return nil, nil
	}

	for _, route := range app.Routes {
		// Parse route string from URL into hostname, domain, and path
		newRoute, err := createRoute(route, space.Name)
		if err != nil {
			return nil, err
		}
		routes = append(routes, newRoute)
	}

	return routes, nil
}

func createAppsToDeploy(pushManifest manifest.Manifest, appName string) ([]manifest.Application, error) {
	appsToDeploy := pushManifest.Applications
	if appName != "" {
		// deploy one app from the manifest
		app, err := pushManifest.App(appName)
		if err != nil {
			return nil, err
		}

		appsToDeploy = []manifest.Application{*app}
	}

	return appsToDeploy, nil
}

func createAppOverrides(cmd *cobra.Command, params *pushParams) (*manifest.Application, error) {

	overrides := &manifest.Application{}
	overrides.Docker.Image = params.containerImage
	overrides.Stack = params.stack
	overrides.Command = params.startupCommand
	overrides.Args = params.containerArgs
	overrides.Entrypoint = params.containerEntrypoint
	overrides.Dockerfile.Path = params.dockerfilePath
	overrides.DiskQuota = params.diskQuota
	overrides.Memory = params.memoryLimit
	overrides.CPU = params.cpu

	// Read environment variables from cli args
	envVars, err := envutil.ParseCLIEnvVars(params.envs)
	if err != nil {
		return nil, err
	}
	overrides.Env = envutil.EnvVarsToMap(envVars)

	if params.buildpack != "" {
		overrides.Buildpacks = []string{params.buildpack}
	}

	overrides.HealthCheckTimeout = params.healthCheckTimeout
	overrides.HealthCheckHTTPEndpoint = params.healthCheckHTTPEndpoint

	if params.healthCheckType != "" {
		overrides.HealthCheckType = params.healthCheckType
	}

	if len(params.rawRoutes) > 0 {
		overrides.Routes = nil
		for _, rr := range params.rawRoutes {
			overrides.Routes = append(overrides.Routes, manifest.Route{
				Route: rr,
			})
		}
	}

	// Only override if the user explicitly set it.
	if cmd.Flags().Lookup("no-route").Changed {
		overrides.NoRoute = &params.noRoute
	}

	// Only override if the user explicitly set it.
	if cmd.Flags().Lookup("random-route").Changed {
		overrides.RandomRoute = &params.randomRouteDomain
	}

	// Only override if the user explicitly set it.
	if cmd.Flags().Lookup("instances").Changed {
		overrides.Instances = &params.instances
	}

	// Only override if the user explicitly set it.
	if cmd.Flags().Lookup("no-start").Changed {
		overrides.NoStart = ptr.Bool(params.noStart)
	}

	// Only override if the user explicitly set it.
	if cmd.Flags().Lookup("task").Changed {
		overrides.Task = &params.task
	}

	return overrides, nil
}

func createManifest(
	ctx context.Context,
	params *pushParams,
	appName string,
) (*manifest.Manifest, error) {
	variables := make(map[string]interface{})

	for _, path := range params.varsFiles {
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("couldn't read var-file %q: %v", path, err)
		}
		if err = yaml.Unmarshal(bytes, &variables); err != nil {
			return nil, fmt.Errorf("invalid var-file %q: %v", path, err)
		}
	}

	for k, v := range params.vars {
		variables[k] = v
	}

	var pushManifest *manifest.Manifest
	var err error
	switch {
	case params.noManifest:
		if pushManifest, err = manifest.New(appName); err != nil {
			return nil, err
		}
	case params.manifestFile != "":
		if pushManifest, err = manifest.NewFromFile(ctx, params.manifestFile, variables); err != nil {
			return nil, fmt.Errorf("supplied manifest file %s resulted in error: %v", params.manifestFile, err)
		}
	default:
		if pushManifest, err = manifest.CheckForManifest(ctx, variables); err != nil {
			return nil, fmt.Errorf("couldn't read manifest file: %v", err)
		}

		if pushManifest == nil {
			if pushManifest, err = manifest.New(appName); err != nil {
				return nil, err
			}
		}
	}

	return pushManifest, nil
}

func pushSourceImage(
	ctx context.Context,
	cmd *cobra.Command,
	b SrcImageBuilder,
	app manifest.Application,
	srcPath string,
	sourceImage string,
) (string, error) {
	// Kontext has to have a absolute path.
	srcPath, err := filepath.Abs(srcPath)
	if err != nil {
		return "", err
	}

	// Sanity check that the Dockerfile is in the source
	if app.Dockerfile.Path != "" {
		absDockerPath := filepath.Join(srcPath, filepath.FromSlash(app.Dockerfile.Path))
		if _, err := os.Stat(absDockerPath); os.IsNotExist(err) {
			logging.FromContext(ctx).Infof("app root: %s", srcPath)
			return "", fmt.Errorf("the Dockerfile %q couldn't be found under the app root", app.Dockerfile.Path)
		}
	}

	filter, err := sourceimage.BuildIgnoreFilter(srcPath)
	if err != nil {
		return "", err
	}

	imageWithDigest, err := b.BuildSrcImage(srcPath, sourceImage, filter)
	if err != nil {
		return "", err
	}

	return imageWithDigest, nil
}

func (p *pushParams) sourceRegistryName(space v1alpha1.Space) string {
	registry := p.containerRegistry
	switch {
	case registry != "":
		break
	default:
		registry = space.Status.BuildConfig.ContainerRegistry
	}

	return registry
}

// checkStackExists ensures that a matching stack is available in the Space if
// specified in the Application.
// NOTE: Currently the return value is ONLY a V3 stack because the stack
// object is ONLY used when ADX Builds are used. Therefore, given ADX Builds
// only supports V3 stacks, we only return V3. This implies when a V2 stack is
// used, a default stack and non-error are returned.
func checkStackExists(
	ctx context.Context,
	p *config.KfParams,
	app manifest.Application,
	space *v1alpha1.Space,
) (kfconfig.StackV3Definition, error) {
	if app.Stack == "" {
		return kfconfig.StackV3Definition{}, nil
	}

	if p.FeatureFlags(ctx).AppDevExperienceBuilds().IsDisabled() {
		for _, stack := range space.Status.BuildConfig.StacksV2 {
			if stack.Name == app.Stack {
				return kfconfig.StackV3Definition{}, nil
			}
		}
	}
	for _, stack := range space.Status.BuildConfig.StacksV3 {
		if stack.Name == app.Stack {
			return stack, nil
		}
	}
	return kfconfig.StackV3Definition{}, fmt.Errorf("no matching stack %q found in space %q", app.Stack, space.Name)
}
