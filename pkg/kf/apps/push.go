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
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"path"
	"strconv"
	"time"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/envutil"
	"github.com/google/kf/v2/pkg/kf/describe"
	"github.com/google/kf/v2/pkg/kf/dynamicutils"
	utils "github.com/google/kf/v2/pkg/kf/internal/utils/cli"
	"github.com/google/kf/v2/pkg/kf/secrets"
	"github.com/google/kf/v2/pkg/kf/serviceinstancebindings"
	"github.com/google/kf/v2/pkg/kf/sourcepackages"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/ptr"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go push-options.yml push_options.go

// pusher deploys source code to Kubernetes. It should be created via
// NewPusher.
type pusher struct {
	appsClient          Client
	bindingsClient      serviceinstancebindings.Client
	secretsClient       secrets.Client
	sourcePackageClient sourcepackages.Client
	poster              sourcepackages.Poster
}

// Pusher deploys applications.
type Pusher interface {
	// Push deploys an application.
	Push(ctx context.Context, appName string, opts ...PushOption) error

	// CreatePlaceholderApp creates a valid stopped application with the given name
	// if the App doesn't exist yet.
	CreatePlaceholderApp(ctx context.Context, appName string, opts ...PushOption) (*v1alpha1.App, error)
}

// NewPusher creates a new Pusher.
func NewPusher(
	appsClient Client,
	bindingsClient serviceinstancebindings.Client,
	secretClient secrets.Client,
	sourcePackageClient sourcepackages.Client,
	poster sourcepackages.Poster,
) Pusher {
	return &pusher{
		appsClient:          appsClient,
		bindingsClient:      bindingsClient,
		secretsClient:       secretClient,
		sourcePackageClient: sourcePackageClient,
		poster:              poster,
	}
}

func newApp(appName, adxBuild string, opts ...PushOption) *v1alpha1.App {

	cfg := PushOptionDefaults().Extend(opts).toConfig()

	var buildRef *corev1.LocalObjectReference
	if adxBuild != "" {
		buildRef = &corev1.LocalObjectReference{
			Name: adxBuild,
		}
	}

	return &v1alpha1.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "kf.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        appName,
			Namespace:   cfg.Space,
			Labels:      cfg.Labels,
			Annotations: cfg.Annotations,
		},
		Spec: v1alpha1.AppSpec{
			Build: v1alpha1.AppSpecBuild{
				Spec:     cfg.Build,
				Image:    cfg.ContainerImage,
				BuildRef: buildRef,
			},
			Template: v1alpha1.AppSpecTemplate{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{cfg.Container},
				},
			},
			Instances: cfg.AppSpecInstances,
			Routes:    cfg.Routes,
		},
	}
}

func newBindings(appName string, opts ...PushOption) ([]v1alpha1.ServiceInstanceBinding, error) {
	cfg := PushOptionDefaults().Extend(opts).toConfig()
	return cfg.ServiceBindings, nil
}

// Push deploys an application to Kubernetes. It can be configured via
// Optionapp.
func (p *pusher) Push(ctx context.Context, appName string, opts ...PushOption) error {
	cfg := PushOptionDefaults().Extend(opts).toConfig()
	logger := logging.FromContext(ctx)

	// Create a stopped App if it doesn't exist.
	app, err := p.CreatePlaceholderApp(ctx, appName, opts...)
	if err != nil {
		return err
	}

	// Bind declared services to the App.
	if err := p.ReconcileBindings(ctx, appName, app, opts...); err != nil {
		return err
	}

	var adxBuild string
	if cfg.ADXBuild {
		// Create the Build for the App to reference.
		buildName, err := createADXBuild(
			ctx,
			cfg.Space,
			app,
			cfg.ADXContainerRegistry,
			cfg.ADXDockerfile,
			cfg.SourcePath,
			cfg.ADXStack,
			p.poster,
		)
		if err != nil {
			return err
		}

		adxBuild = buildName
	}

	app = newApp(appName, adxBuild, opts...)

	var hasDefaultRoutes bool
	app.Spec.Routes, hasDefaultRoutes = setupRoutes(cfg, app.Name, app.Spec.Routes)

	// Scaling
	if noScaling(app.Spec.Instances) {
		// Default to 1
		app.Spec.Instances.Replicas = ptr.Int32(1)
	}

	// If there is a source path set, then we need to setup the App's Build to
	// look for a SourcePackage.
	if cfg.SourcePath != "" && !cfg.ADXBuild {
		if app.Spec.Build.Spec == nil {
			app.Spec.Build.Spec = &v1alpha1.BuildSpec{}
		}

		app.Spec.Build.Spec.SourcePackage.Name = sourcePackageName(app)
	}

	resultingApp, err := p.appsClient.Upsert(
		ctx,
		app.Namespace,
		app,
		mergeApps(cfg, hasDefaultRoutes),
	)
	if err != nil {
		return fmt.Errorf("failed to push App: %s", err)
	}

	// Upload the source directory.
	if cfg.SourcePath != "" && !cfg.ADXBuild {
		if err := p.sourcePackageClient.UploadSourcePath(
			ctx,
			cfg.SourcePath,
			resultingApp,
		); err != nil {
			return err
		}
	}

	if err := p.appsClient.DeployLogsForApp(ctx, cfg.Output, resultingApp); err != nil {
		return err
	}

	status := "deployed"
	if resultingApp.Spec.Instances.Stopped {
		status = "deployed without starting"

		utils.SuggestNextAction(utils.NextAction{
			Description: "Start App",
			Commands: []string{
				fmt.Sprintf("kf start %s --space %s", resultingApp.Name, resultingApp.Namespace),
			},
		})
	} else {
		// Give the user a way to back out non-destructively
		utils.SuggestNextAction(utils.NextAction{
			Description: "Stop App",
			Commands: []string{
				fmt.Sprintf("kf stop %s --space %s", resultingApp.Name, resultingApp.Namespace),
			},
		})
	}

	logger.Infof("Successfully %s", status)

	finalApp, err := p.appsClient.Get(ctx, app.Namespace, app.Name)
	if err != nil {
		return fmt.Errorf("App deployed, but couldn't fetch final status: %s", err)
	}

	var startCmd []string

	if finalApp.Status.StartCommands.Buildpack != nil {
		startCmd = finalApp.Status.StartCommands.Buildpack
	} else {
		startCmd = finalApp.Status.StartCommands.Container
	}

	// Show enough info to start using the App.
	describe.TabbedWriter(cfg.Output, func(w io.Writer) {
		fmt.Fprintf(w, "Name:\t%s\n", finalApp.Name)
		fmt.Fprintf(w, "Space:\t%s\n", finalApp.Namespace)
		fmt.Fprintf(w, "Routes:\t%v\n", finalApp.Status.URLs)
		fmt.Fprintf(w, "Instances:\t%s\n", finalApp.Status.Instances.Representation)
		fmt.Fprintf(w, "Start command:\t%v\n", startCmd)
	})

	return nil
}

// CreatePlaceholderApp creates a valid stopped application with the given name
// if the App doesn't exist yet.
func (p *pusher) CreatePlaceholderApp(ctx context.Context, appName string, opts ...PushOption) (*v1alpha1.App, error) {
	cfg := PushOptionDefaults().Extend(opts).toConfig()
	logger := logging.FromContext(ctx)

	logger.Infof("Checking for existing App named %q...", appName)
	app, err := p.appsClient.Get(ctx, cfg.Space, appName)
	switch {
	case apierrs.IsNotFound(err):
		logger.Infof("App %q doesn't exist, creating placeholder...", appName)

		// If the App doesn't exist, create a placeholder.
		tmpApp := newApp(
			appName,
			"",
			WithPushSpace(cfg.Space),
			WithPushContainerImage(ptr.String("gcr.io/kf-releases/nop:nop")),
			WithPushAppSpecInstances(
				v1alpha1.AppSpecInstances{
					Stopped: true,
				},
			),
			WithPushLabels(cfg.Labels),
			WithPushAnnotations(cfg.Annotations),
		)

		app, err := p.appsClient.Create(ctx, cfg.Space, tmpApp)
		if err != nil {
			return nil, fmt.Errorf("couldn't create App placeholder: %v", err)
		}

		// Wait for ready.
		if _, err := p.appsClient.WaitForConditionReadyTrue(ctx, cfg.Space, appName, 1*time.Second); err != nil {
			return nil, fmt.Errorf("couldn't wait for App placeholder: %v", err)
		}
		logger.Info("Placeholder App created.")
		return app, nil

	case err != nil:
		return nil, fmt.Errorf("failed to get App %q: %v", appName, err)

	default:
		logger.Infof("App %q exists, it will be updated.", appName)
		return app, nil
	}
}

// ReconcileBindings binds services declared in a manifest to the given App.
func (p *pusher) ReconcileBindings(ctx context.Context, appName string, app *v1alpha1.App, opts ...PushOption) error {
	cfg := PushOptionDefaults().Extend(opts).toConfig()
	logger := logging.FromContext(ctx)

	bindings := cfg.ServiceBindings
	logger.Infof("Binding %d services to App...", len(bindings))

	// Create service instance bindings
	for _, desiredBinding := range bindings {
		logger.Infof("Checking for binding to ServiceInstance named %q...", desiredBinding.Spec.InstanceRef.Name)
		_, err := p.bindingsClient.Get(ctx, desiredBinding.GetNamespace(), desiredBinding.Name)
		switch {
		case apierrs.IsNotFound(err):
			logger.Infof("Creating binding for ServiceInstance %q...", desiredBinding.Spec.InstanceRef.Name)

			desiredInstanceBindingOwnerReferences := []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			}

			desiredBinding.ObjectMeta.OwnerReferences = desiredInstanceBindingOwnerReferences

			newBinding, err := p.bindingsClient.Create(ctx, desiredBinding.GetNamespace(), &desiredBinding)

			if err != nil {
				return fmt.Errorf("failed to create ServiceInstanceBinding: %s", err)
			}

			_, err = p.secretsClient.CreateParamsSecret(ctx, newBinding, newBinding.Spec.ParametersFrom.Name, json.RawMessage("{}"))
			if err != nil {
				return fmt.Errorf("failed to create binding parameters secret: %s", err)
			}

			// Wait for binding to become ready
			logger.Info("Waiting for ServiceInstanceBinding to become ready")
			if _, err = p.bindingsClient.WaitForConditionReadyTrue(
				context.Background(), newBinding.GetNamespace(), newBinding.Name, 1*time.Second); err != nil {
				return fmt.Errorf("service binding failed: %s", err)
			}
			logger.Info("Success")
		case err != nil:
			return fmt.Errorf("failed to get ServiceInstanceBinding %q: %s", desiredBinding.Name, err)
		default:
			logger.Infof("Binding to ServiceInstance %q already exists.\n", desiredBinding.Spec.InstanceRef.Name)
		}
	}

	logger.Info("Finished binding services.")
	return nil
}

func setupRoutes(cfg pushConfig, appName string, r []v1alpha1.RouteWeightBinding) (routes []v1alpha1.RouteWeightBinding, hasDefaultRoutes bool) {
	switch {
	case len(r) != 0:
		// Don't overwrite the routes
		return r, false
	case cfg.GenerateDefaultRoute:
		return []v1alpha1.RouteWeightBinding{
			{
				RouteSpecFields: v1alpha1.RouteSpecFields{
					Hostname: appName,
				},
			},
		}, true
	case cfg.GenerateRandomRoute:
		return []v1alpha1.RouteWeightBinding{
			{
				RouteSpecFields: v1alpha1.RouteSpecFields{
					Hostname: v1alpha1.GenerateName(
						appName,
						strconv.FormatUint(rand.Uint64(), 36),
						strconv.FormatUint(uint64(time.Now().UnixNano()), 36),
					),
				},
			},
		}, true
	default:
		return nil, false
	}
}

func noScaling(instances v1alpha1.AppSpecInstances) bool {
	return instances.Replicas == nil
}

func sourcePackageName(app *v1alpha1.App) string {
	return v1alpha1.GenerateName(
		app.Name,
		strconv.FormatInt(int64(app.Spec.Build.UpdateRequests), 10),
	)
}

func mergeApps(cfg pushConfig, hasDefaultRoutes bool) func(newapp, oldapp *v1alpha1.App) *v1alpha1.App {
	return func(newapp, oldapp *v1alpha1.App) *v1alpha1.App {
		// UpdateRequests
		// Always increment to ensure the app is redeployed. This is
		// especially important if we are not creating a new Build which has
		// its own UpdateRequests field.
		newapp.Spec.Template.UpdateRequests = oldapp.Spec.Template.UpdateRequests + 1

		// Always increment to ensure a new Build is created if necessary.
		newapp.Spec.Build.UpdateRequests = oldapp.Spec.Build.UpdateRequests + 1

		// Have the SourcePackage name match the Build's (and its new
		// UpdateRequests).
		if newapp.Spec.Build.Spec != nil && newapp.Spec.Build.Spec.SourcePackage.Name != "" {
			newapp.Spec.Build.Spec.SourcePackage.Name = sourcePackageName(newapp)
		}

		// Routes
		if len(oldapp.Spec.Routes) > 0 && hasDefaultRoutes {
			newapp.Spec.Routes = oldapp.Spec.Routes
		}

		// Scaling overrides
		if noScaling(cfg.AppSpecInstances) {
			// Looks like the user did not set a new value, use the old one
			newapp.Spec.Instances.Replicas = oldapp.Spec.Instances.Replicas
		}

		// Default scaling
		if noScaling(cfg.AppSpecInstances) && noScaling(oldapp.Spec.Instances) {
			// No scaling in old or new, go with a default of 1. This is to
			// match expectaions for CF users. See
			// https://github.com/google/kf/v2/issues/8 for more context.
			newapp.Spec.Instances.Replicas = ptr.Int32(1)
		}

		newapp.ResourceVersion = oldapp.ResourceVersion

		// Envs
		newEnvs := envutil.GetAppEnvVars(newapp)
		oldEnvs := envutil.GetAppEnvVars(oldapp)
		envutil.SetAppEnvVars(
			newapp,
			envutil.DeduplicateEnvVars(append(oldEnvs, newEnvs...)),
		)

		// Resource merging
		for k, v := range getResources(newapp) {
			setResourceRequest(oldapp, k, v)
		}
		setResources(newapp, getResources(oldapp))

		return newapp
	}
}

func getResources(app *v1alpha1.App) corev1.ResourceList {
	containers := app.Spec.Template.Spec.Containers

	if len(containers) == 0 {
		return nil
	}

	return containers[0].Resources.Requests
}

func setResources(app *v1alpha1.App, l corev1.ResourceList) {
	containers := app.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return
	}

	containers[0].Resources.Requests = l
}

func setResourceRequest(app *v1alpha1.App, name corev1.ResourceName, v resource.Quantity) {
	containers := app.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return
	}

	if containers[0].Resources.Requests == nil {
		containers[0].Resources.Requests = corev1.ResourceList{}
	}

	containers[0].Resources.Requests[name] = v
}

// SourceImageName gets the image name for source code for an application.
func SourceImageName(namespace, appName string) string {
	return fmt.Sprintf("src-%s-%s", namespace, appName)
}

// JoinRepositoryImage joins a repository and image name.
func JoinRepositoryImage(repository, imageName string) string {
	return path.Join(repository, imageName)
}

func createADXBuild(
	ctx context.Context,
	namespace string,
	app *v1alpha1.App,
	containerRegistry string,
	dockerfilePath string,
	sourcePath string,
	stack config.StackV3Definition,
	poster sourcepackages.Poster,
) (buildName string, err error) {
	adxBuildsClient := dynamicclient.
		Get(ctx).
		Resource(schema.GroupVersionResource{
			Group:    "builds.appdevexperience.dev",
			Version:  "v1alpha1",
			Resource: "builds",
		}).
		Namespace(namespace)
	adxSPClient := dynamicclient.
		Get(ctx).
		Resource(schema.GroupVersionResource{
			Group:    "builds.appdevexperience.dev",
			Version:  "v1alpha1",
			Resource: "sourcepackages",
		}).
		Namespace(namespace)

	// NOTE: We use a random number formatted in base 64 (0-9-a-z) to
	// allow for a smaller number of characters to make up the tag.
	outputName := JoinRepositoryImage(containerRegistry,
		fmt.Sprintf("%s-%s:%s", namespace, app.Name, strconv.FormatUint(uint64(rand.Uint32()), 36)),
	)
	if len(outputName) >= 100 {
		// If docker names end up longer than 100 characters, it can bump into
		// issues: https://github.com/docker/for-linux/issues/484
		outputName = JoinRepositoryImage(containerRegistry, strconv.FormatUint(uint64(rand.Uint64()), 36))
	}

	// Package up the path.
	cleanup, tarFile, cz, err := sourcepackages.PackageSourcePath(sourcePath)

	// Always invoke the cleanup (even when there is an error). This will
	// delete the temp tar file.
	defer cleanup()

	if err != nil {
		return "", fmt.Errorf("failed to package source path: %v", err)
	}

	// Get the size of the tar file.
	fi, err := tarFile.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info for tar file: %v", err)
	}

	// Create the SourcePackage for the Build to reference.
	sourcePackage := dynamicutils.NewUnstructured(map[string]interface{}{
		"apiVersion":            "builds.appdevexperience.dev/v1alpha1",
		"kind":                  "SourcePackage",
		"metadata.generateName": app.Name + "-",
		"metadata.namespace":    namespace,
		"metadata.ownerReferences": []metav1.OwnerReference{
			*kmeta.NewControllerRef(app),
		},
		"metadata.labels": map[string]string{
			v1alpha1.NameLabel: app.Name,
		},
		"spec.size":           uint64(fi.Size()),
		"spec.checksum.type":  "sha256",
		"spec.checksum.value": cz,
	})
	actualSpU, err := adxSPClient.Create(ctx, sourcePackage, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create SourcePackage: %v", err)
	}

	buildM := map[string]interface{}{
		"apiVersion":            "builds.appdevexperience.dev/v1alpha1",
		"kind":                  "Build",
		"metadata.generateName": app.Name + "-",
		"metadata.namespace":    namespace,
		"metadata.ownerReferences": []metav1.OwnerReference{
			*kmeta.NewControllerRef(app),
		},
		"metadata.labels": map[string]string{
			v1alpha1.NameLabel: app.Name,
		},
		"spec.outputImage":                   outputName,
		"spec.sourceCode.sourcePackage.name": actualSpU.GetName(),
	}

	// Set values from the stack if they are set.
	if dockerfilePath != "" {
		// Use dockerfile build.
		buildM["spec.dockerfileSpec.dockerfile"] = dockerfilePath
	} else {
		// Use buildpacks build.
		if stack.BuildImage != "" {
			buildM["spec.buildpackSpec.builder"] = stack.BuildImage
		}
		if stack.RunImage != "" {
			buildM["spec.buildpackSpec.runImage"] = stack.RunImage
		}
		if len(stack.NodeSelector) != 0 {
			buildM["spec.runtime.nodeSelector"] = stack.NodeSelector
		}
	}

	build := dynamicutils.NewUnstructured(buildM)

	actualU, err := adxBuildsClient.Create(ctx, build, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create Build: %v", err)
	}

	if err := poster(ctx,
		fmt.Sprintf(
			"/apis/data.builds.appdevexperience.dev/v1alpha1/proxy/namespaces/%s/%s/source",
			app.Namespace,
			actualSpU.GetName(),
		),
		tarFile.Name(),
	); err != nil {
		return "", fmt.Errorf("failed to post tar file: %v", err)
	}

	return actualU.GetName(), nil
}
