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

package v1alpha1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/ptr"
)

var (
	defaultMem     = resource.MustParse("1Gi")
	defaultStorage = resource.MustParse("1Gi")

	// CPU isn't defaulted in Cloud Foundry so we assume apps are I/O bound
	// and roughly 10 should run on a single core machine.
	defaultCPU = resource.MustParse("100m")
)

const (
	// DefaultHealthCheckProbeTimeout holds the default timeout to be applied to
	// healthchecks in seconds. This matches Cloud Foundry's default timeout.
	DefaultHealthCheckProbeTimeout = 60

	// DefaultHealthCheckProbeEndpoint is the default endpoint to use for HTTP
	// Get health checks.
	DefaultHealthCheckProbeEndpoint = "/"

	// DefaultHealthCheckPeriodSeconds holds the default period between health
	// check polls.
	DefaultHealthCheckPeriodSeconds = 10

	// DefaultHealthCheckFailureThreshold is the number of times the probe can
	// return a failure before the app is considered bad.
	DefaultHealthCheckFailureThreshold = 3
)

// SetDefaults implements apis.Defaultable
func (k *App) SetDefaults(ctx context.Context) {
	k.Spec.SetDefaults(ctx)
}

// SetDefaults implements apis.Defaultable
func (k *AppSpec) SetDefaults(ctx context.Context) {
	k.SetBuildDefaults(ctx)
	k.Template.SetDefaults(ctx, k)
	k.Instances.SetDefaults(ctx)
	k.SetRouteDefaults(ctx)
}

// SetBuildDefaults implements apis.Defaultable for the embedded BuildSpec.
func (k *AppSpec) SetBuildDefaults(ctx context.Context) {

	// If the app build has changed without changing the UpdateRequests,
	// update it.
	if base := apis.GetBaseline(ctx); base != nil {
		if old, ok := base.(*App); ok {
			// If the update is a post rather than a patch, pick up where the last
			// build left off.
			if k.Build.UpdateRequests == 0 {
				k.Build.UpdateRequests = old.Spec.Build.UpdateRequests
			}

			if k.Build.NeedsUpdateRequestsIncrement(old.Spec.Build) {
				k.Build.UpdateRequests++
			}
		}
	}
}

// SetRouteDefaults sets the defaults for an AppSpec's Routes.
func (k *AppSpec) SetRouteDefaults(ctx context.Context) {
	for i := range k.Routes {
		route := &k.Routes[i]
		route.SetDefaults(ctx)
	}

	// Deduplicate routes with default properties.
	k.Routes = MergeBindings(k.Routes)
}

// SetDefaults implements apis.Defaultable.
func (k *AppSpecTemplate) SetDefaults(ctx context.Context, spec *AppSpec) {

	// If the app has changed without changing the UpdateRequests, update it.
	if base := apis.GetBaseline(ctx); base != nil {
		if old, ok := base.(*App); ok {
			// If the update is a post rather than a patch, pick up where the
			// last template left off.
			if k.UpdateRequests == 0 {
				k.UpdateRequests = old.Spec.Template.UpdateRequests
			}

			if spec.NeedsUpdateRequestsIncrement(old.Spec) {
				k.UpdateRequests++
			}
		}
	}

	// We require at least one container, so if there isn't one, set a blank
	// one.
	if len(k.Spec.Containers) == 0 {
		k.Spec.Containers = append(k.Spec.Containers, corev1.Container{})
	}

	container := &k.Spec.Containers[0]
	SetKfAppContainerDefaults(ctx, container)
}

// SetDefaults implements apis.Defaultable.
func (k *AppSpecInstances) SetDefaults(ctx context.Context) {
	defer func() {
		k.DeprecatedExactly = nil
	}()

	switch {
	case k.Replicas != nil:
		// Replicas already set, move on.
	case k.DeprecatedExactly == nil:
		// Nothing is set, default to 1
		k.Replicas = ptr.Int32(1)
	default:
		// Promote DeprecatedExactly
		k.Replicas = k.DeprecatedExactly
	}

	k.Autoscaling.SetAutoscalingDefaults(ctx)
}

// SetAutoscalingDefaults set the defaults for AppSpecAutoscaling
func (as *AppSpecAutoscaling) SetAutoscalingDefaults(_ context.Context) {
	if as.MaxReplicas != nil && as.MinReplicas == nil {
		as.MinReplicas = ptr.Int32(1)
	}
}

// SetDefaults implements apis.Defaultable.
func (s *Scale) SetDefaults(ctx context.Context) {
	// No defaults
}

// SetKfAppContainerDefaults sets the defaults for an application container.
// This function MAY be context sensitive in the future.
func SetKfAppContainerDefaults(_ context.Context, container *corev1.Container) {
	if container.Name == "" {
		container.Name = DefaultUserContainerName
	}

	setContainerReadinessProbe(container)

	// Set default disk, RAM, and CPU limits on the application if they have not been custom set
	if container.Resources.Requests == nil {
		container.Resources.Requests = v1.ResourceList{}
	}

	if _, exists := container.Resources.Requests[corev1.ResourceMemory]; !exists {
		container.Resources.Requests[corev1.ResourceMemory] = defaultMem
	}

	if _, exists := container.Resources.Requests[corev1.ResourceEphemeralStorage]; !exists {
		container.Resources.Requests[corev1.ResourceEphemeralStorage] = defaultStorage
	}

	if _, exists := container.Resources.Requests[corev1.ResourceCPU]; !exists {
		container.Resources.Requests[corev1.ResourceCPU] = defaultCPU
	}
}

func setContainerReadinessProbe(container *corev1.Container) {
	// Container.ReadinessProbe is set at client side
	// based on App's health check type (http/port/process/none).
	// If container.ReadinessProbe == nil, it means health check type == "process" or "none",
	// it shouldn't be default at the Webhook where App's health check type can't be determined,
	// it could result in incorrect configuration if setting container.ReadinessProbe for "process"
	// or "none" health check types (see b/173615950).
	if container.ReadinessProbe == nil {
		return
	}

	readinessProbe := container.ReadinessProbe

	// Default the probe timeout
	if readinessProbe.TimeoutSeconds == 0 {
		readinessProbe.TimeoutSeconds = DefaultHealthCheckProbeTimeout
	}

	// Knative serving 0.8 requires PeriodSeconds and FailureThreshold to be set
	// even if we don't expose them to users directly.
	if readinessProbe.PeriodSeconds == 0 {
		readinessProbe.PeriodSeconds = DefaultHealthCheckPeriodSeconds
	}

	if readinessProbe.FailureThreshold == 0 {
		readinessProbe.FailureThreshold = DefaultHealthCheckFailureThreshold
	}

	// If the probe is HTTP, default the path
	if http := readinessProbe.HTTPGet; http != nil {
		if http.Path == "" {
			http.Path = DefaultHealthCheckProbeEndpoint
		}
	}
}
