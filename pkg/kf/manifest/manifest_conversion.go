// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manifest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/internal/envutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ToAppSpecInstances extracts scaling info from the manifest.
func (source *Application) ToAppSpecInstances() v1alpha1.AppSpecInstances {
	instances := v1alpha1.AppSpecInstances{}
	if source.NoStart != nil {
		instances.Stopped = *source.NoStart
	}

	instances.Min = source.MinScale
	instances.Max = source.MaxScale
	instances.Exactly = source.Instances

	return instances
}

// intPtr creates a pointer to an int.
func intPtr(i int) *int {
	return &i
}

// ToResourceRequests returns a ResourceList with memory, CPU, and storage set.
// If none are set by the user, the returned ResourceList will be nil.
func (source *Application) ToResourceRequests() (corev1.ResourceList, error) {
	resourceMapping := map[corev1.ResourceName]string{
		corev1.ResourceMemory:           cfToSIUnits(source.Memory),
		corev1.ResourceEphemeralStorage: cfToSIUnits(source.DiskQuota),
		// CPU is not converted to SI because it's not a normal CF field
		// and is therefore expected to be in SI to begin with.
		corev1.ResourceCPU: source.CPU,
	}

	requests := corev1.ResourceList{}
	for kind, rawQuantity := range resourceMapping {
		if rawQuantity != "" {
			quantity, err := resource.ParseQuantity(rawQuantity)
			if err != nil {
				return nil, fmt.Errorf("couldn't parse resource quantity %s: %v", rawQuantity, err)
			}

			requests[kind] = quantity
		}
	}

	if len(requests) == 0 {
		return nil, nil
	}

	return requests, nil
}

// cfToSiUnits converts CF resource quantities into the equivalent k8s quantity
// strings. CF interprets K, M, G, T as binary SI units while k8s interprets
// them as decimal, so we convert them here into binary SI units (Ki, Mi, Gi, Ti)
func cfToSIUnits(orig string) string {
	trimmed := strings.TrimSuffix(strings.TrimSpace(orig), "B")

	for _, suffix := range []string{"T", "G", "M", "K"} {
		if strings.HasSuffix(trimmed, suffix) {
			return trimmed + "i"
		}
	}

	// if it's not a CF unit, return the value unmodified
	return orig
}

// ToHealthCheck creates a corev1.Probe that maps the health checks CloudFoundry
// does.
func (source *Application) ToHealthCheck() (*corev1.Probe, error) {
	if source.HealthCheckTimeout < 0 {
		return nil, errors.New("health check timeouts can't be negative")
	}

	probe := &corev1.Probe{TimeoutSeconds: int32(source.HealthCheckTimeout)}

	switch source.HealthCheckType {
	case "http":
		probe.Handler.HTTPGet = &corev1.HTTPGetAction{Path: source.HealthCheckHTTPEndpoint}
		return probe, nil

	case "port", "": // By default, cf uses a port based health check.
		if source.HealthCheckHTTPEndpoint != "" {
			return nil, errors.New("health check endpoints can only be used with http checks")
		}

		probe.Handler.TCPSocket = &corev1.TCPSocketAction{}
		return probe, nil

	case "process", "none":
		return nil, errors.New("kf doesn't support the process health check type")

	default:
		return nil, fmt.Errorf("unknown health check type %s, supported types are http and port", source.HealthCheckType)
	}
}

// ToContainer converts the manifest to a container suitable for use in a pod,
// ksvc, or app.
func (source *Application) ToContainer() (corev1.Container, error) {
	resourceRequests, err := source.ToResourceRequests()
	if err != nil {
		return corev1.Container{}, err
	}

	healthCheck, err := source.ToHealthCheck()
	if err != nil {
		return corev1.Container{}, err
	}

	container := corev1.Container{
		Args:    source.CommandArgs(),
		Command: source.CommandEntrypoint(),
		Resources: corev1.ResourceRequirements{
			Requests: resourceRequests,
		},
		ReadinessProbe: healthCheck,
	}

	if len(source.Env) > 0 {
		container.Env = envutil.MapToEnvVars(source.Env)
	}

	if source.EnableHTTP2 != nil && *source.EnableHTTP2 {
		container.Ports = HTTP2ContainerPort()
	}

	return container, nil
}

// HTTP2ContainerPort returns the container port used to notify Knative that
// the container should be allowed to accept HTTP2 requests.
func HTTP2ContainerPort() []corev1.ContainerPort {
	return []corev1.ContainerPort{{Name: "h2c", ContainerPort: 8080}}
}
