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
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/apis"
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
	k.SetSourceDefaults(ctx)
	k.Template.SetDefaults(ctx)
	k.SetServiceBindingDefaults(ctx)
}

// SetSourceDefaults implements apis.Defaultable for the embedded SourceSpec.
func (k *AppSpec) SetSourceDefaults(ctx context.Context) {

	// If the app source has changed without changing the UpdateRequests,
	// update it.
	if base := apis.GetBaseline(ctx); base != nil {
		if old, ok := base.(*App); ok {
			// If the update is a post rather than a patch, pick up where the last
			// source left off.
			if k.Source.UpdateRequests == 0 {
				k.Source.UpdateRequests = old.Spec.Source.UpdateRequests
			}

			if k.Source.NeedsUpdateRequestsIncrement(old.Spec.Source) {
				k.Source.UpdateRequests++
			}
		}
	}
}

// SetServiceBindingDefaults sets the defaults for an AppSpec's ServiceBindings.
func (k *AppSpec) SetServiceBindingDefaults(ctx context.Context) {
	for i := range k.ServiceBindings {
		binding := &k.ServiceBindings[i]
		binding.SetDefaults(ctx)
	}
}

// SetDefaults sets the defaults for an AppSpecServiceBinding.
func (k *AppSpecServiceBinding) SetDefaults(ctx context.Context) {
	if k.BindingName == "" {
		k.BindingName = k.Instance
	}
	if string(k.Parameters) == "" {
		k.Parameters = json.RawMessage("null")
	}
}

// SetDefaults implements apis.Defaultable
func (k *AppSpecTemplate) SetDefaults(ctx context.Context) {

	// We require at least one container, so if there isn't one, set a blank
	// one.
	if len(k.Spec.Containers) == 0 {
		k.Spec.Containers = append(k.Spec.Containers, corev1.Container{})
	}

	container := &k.Spec.Containers[0]
	SetKfAppContainerDefaults(ctx, container)
}

// SetKfAppContainerDefaults sets the defaults for an application container.
// This function MAY be context sensitive in the future.
func SetKfAppContainerDefaults(_ context.Context, container *corev1.Container) {
	if container.Name == "" {
		container.Name = "user-container"
	}

	// Default the probe to a TCP connection if unspecified
	if container.ReadinessProbe == nil {
		container.ReadinessProbe = &corev1.Probe{
			SuccessThreshold: 1,
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{},
			},
		}
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
