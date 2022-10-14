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
	"math"
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/envutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	ramDivisor = resource.MustParse("1Gi")
)

// ToAppSpecInstances extracts scaling info from the manifest.
func (source *Application) ToAppSpecInstances() v1alpha1.AppSpecInstances {
	instances := v1alpha1.AppSpecInstances{}
	if source.NoStart != nil {
		instances.Stopped = *source.NoStart
	}

	if source.Task != nil && *source.Task {
		instances.Stopped = true
	}

	instances.Replicas = source.Instances

	return instances
}

// ToResourceRequirements returns a ResourceRequirements with memory, CPU, and storage set.
func (source *Application) ToResourceRequirements(runtimeConfig *v1alpha1.SpaceStatusRuntimeConfig) (*corev1.ResourceRequirements, error) {
	requirements := &corev1.ResourceRequirements{
		Requests: make(corev1.ResourceList),
		Limits:   make(corev1.ResourceList),
	}

	memoryRequest := v1alpha1.DefaultMem
	if rawMem := CFToSIUnits(source.Memory); rawMem != "" {
		quantity, err := resource.ParseQuantity(rawMem)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse memory %s: %v", rawMem, err)
		}

		requirements.Requests[corev1.ResourceMemory] = quantity
		requirements.Limits[corev1.ResourceMemory] = quantity
		memoryRequest = quantity
	}

	if rawStorage := CFToSIUnits(source.DiskQuota); rawStorage != "" {
		quantity, err := resource.ParseQuantity(rawStorage)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse disk %s: %v", rawStorage, err)
		}

		requirements.Requests[corev1.ResourceEphemeralStorage] = quantity
		requirements.Limits[corev1.ResourceEphemeralStorage] = quantity
	}

	// CPU is not converted to SI because it's not a normal CF field
	// and is therefore expected to be in SI to begin with.
	if rawCPU := source.CPU; rawCPU != "" {
		quantity, err := resource.ParseQuantity(rawCPU)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse cpu %s: %v", rawCPU, err)
		}

		requirements.Requests[corev1.ResourceCPU] = quantity
	} else {
		// If CPU isn't set, we use a CF-ism which is to default it based on the amount of RAM.
		if cpuMult := runtimeConfig.AppCPUPerGBOfRAM; cpuMult != nil {
			ramRatio := float64(memoryRequest.MilliValue()) / float64(ramDivisor.MilliValue())

			cpuMillis := math.Ceil(cpuMult.AsApproximateFloat64() * ramRatio * 1000)
			requirements.Requests[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(cpuMillis), resource.BinarySI)
		}
	}

	// Set a lower-bound on CPU so containers aren't starved.
	if minCPU := runtimeConfig.AppCPUMin; minCPU != nil {
		if cpuRequest, ok := requirements.Requests[corev1.ResourceCPU]; !ok || cpuRequest.Cmp(*minCPU) < 0 {
			requirements.Requests[corev1.ResourceCPU] = *minCPU
		}
	}

	if rawCPULimit := source.CPULimit; rawCPULimit != "" {
		quantity, err := resource.ParseQuantity(rawCPULimit)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse cpu-limit %s: %v", rawCPULimit, err)
		}

		if cpuRequest, ok := requirements.Requests[corev1.ResourceCPU]; ok && quantity.Cmp(cpuRequest) < 0 {
			return nil, fmt.Errorf("cpu-limit: %q must be greater than request: %q", quantity.String(), cpuRequest.String())
		}

		requirements.Limits[corev1.ResourceCPU] = quantity
	}

	if len(requirements.Requests) == 0 {
		requirements.Requests = nil
	}

	if len(requirements.Limits) == 0 {
		requirements.Limits = nil
	}

	return requirements, nil
}

// CFToSIUnits converts CF resource quantities into the equivalent k8s quantity
// strings. CF interprets K, M, G, T as binary SI units while k8s interprets
// them as decimal, so we convert them here into binary SI units (Ki, Mi, Gi, Ti)
func CFToSIUnits(orig string) string {
	trimmed := strings.TrimSuffix(strings.TrimSpace(orig), "B")

	for _, suffix := range []string{"T", "G", "M", "K"} {
		if strings.HasSuffix(trimmed, suffix) {
			return trimmed + "i"
		}
	}

	// if it's not a CF unit, return the value unmodified
	return orig
}

// ToStartupHealthCheck creates a corev1.Probe that mimics Cloud Foundry's
// post-startup application health checks. The following are steps 2 and 3 from
// https://docs.cloudfoundry.org/devguide/deploy-apps/healthchecks.html#healthcheck-lifecycle
//
// When deploying the app, the developer specifies a health check type for the app and,
// optionally, a timeout. If the developer does not specify a health check type, then
// the monitoring process defaults to a port health check.
//
// When Diego starts an app instance, the app health check runs every two seconds
// until aresponse indicates that the app instance is healthy or until the health
// check timeout elapses. The 2-second health check interval is not configurable.
func (source *Application) ToStartupHealthCheck() (*corev1.Probe, error) {
	if source.HealthCheckTimeout < 0 {
		return nil, errors.New("health check timeouts can't be negative")
	}
	healthCheckTimeout := source.HealthCheckTimeout
	if healthCheckTimeout == 0 {

		// https://docs.cloudfoundry.org/devguide/deploy-apps/healthchecks.html#health_check_timeout
		// In Cloud Foundry, the default timeout is 60 seconds
		healthCheckTimeout = 60
	}

	probe := &corev1.Probe{
		// For both HTTP and port based health checks, CF docs state:
		// The configured endpoint must respond within one second to be considered healthy.
		// This was later revised to be user-configurable.
		TimeoutSeconds:   1,
		SuccessThreshold: 1,
		PeriodSeconds:    2,
		ProbeHandler:     corev1.ProbeHandler{},
	}

	if source.HealthCheckInvocationTimeout != 0 {
		probe.TimeoutSeconds = int32(source.HealthCheckInvocationTimeout)
	}

	// To get the startup probe to work like CF where timeout is the total
	// amount of time from container start to health we set the failure threshold
	// to be:
	//
	//     ceil(timeout / check period)

	probe.FailureThreshold = int32(math.Ceil(float64(healthCheckTimeout) / float64(probe.PeriodSeconds)))

	switch source.HealthCheckType {
	case "http":
		probe.ProbeHandler.HTTPGet = &corev1.HTTPGetAction{Path: source.HealthCheckHTTPEndpoint}
		return probe, nil

	case "port", "": // By default, cf uses a port based health check.
		if source.HealthCheckHTTPEndpoint != "" {
			return nil, errors.New("health check endpoints can only be used with http checks")
		}

		probe.ProbeHandler.TCPSocket = &corev1.TCPSocketAction{}
		return probe, nil

	case "process", "none":
		// A process check implies there isn't a probe but instead just rely
		// on the process failing.
		// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#container-probes
		return nil, nil

	default:
		return nil, fmt.Errorf("unknown health check type %s", source.HealthCheckType)
	}
}

// ToPostStartupHealthCheck creates a corev1.Probe that mimics Cloud Foundry's
// post-startup application health checks (used for both liveness and readiness).
// The following are steps 6 and 7 from
// https://docs.cloudfoundry.org/devguide/deploy-apps/healthchecks.html#healthcheck-lifecycle
//
// When an app instance becomes healthy, its route is advertised, if applicable.
// Subsequent health checks are run every 30 seconds once the app becomes healthy.
// The 30-second health check interval is not configurable.
//
// If a previously healthy app instance fails a health check, Diego considers that
// particular instance to be unhealthy. As a result, Diego stops and deletes the
// app instance.
func (source *Application) ToPostStartupHealthCheck() (*corev1.Probe, error) {
	if source.HealthCheckInvocationTimeout < 0 {
		return nil, errors.New("health check invocation timeouts can't be negative")
	}

	probe := &corev1.Probe{
		// For both HTTP and port based health checks, CF docs state:
		// The configured endpoint must respond within one second to be considered healthy.
		// This was later revised to be user-configurable.
		TimeoutSeconds:   1,
		SuccessThreshold: 1,
		FailureThreshold: 1,
		PeriodSeconds:    30,
		ProbeHandler:     corev1.ProbeHandler{},
	}

	if source.HealthCheckInvocationTimeout != 0 {
		probe.TimeoutSeconds = int32(source.HealthCheckInvocationTimeout)
	}

	switch source.HealthCheckType {
	case "http":
		probe.ProbeHandler.HTTPGet = &corev1.HTTPGetAction{Path: source.HealthCheckHTTPEndpoint}
		return probe, nil

	case "port", "": // By default, cf uses a port based health check.
		if source.HealthCheckHTTPEndpoint != "" {
			return nil, errors.New("health check endpoints can only be used with http checks")
		}

		probe.ProbeHandler.TCPSocket = &corev1.TCPSocketAction{}
		return probe, nil

	case "process", "none":
		// A process check implies there isn't a probe but instead just rely
		// on the process failing.
		// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#container-probes
		return nil, nil

	default:
		return nil, fmt.Errorf("unknown health check type %s", source.HealthCheckType)
	}
}

func (source *Application) hasCFHealthCheckFields() bool {
	return source.HealthCheckTimeout != 0 ||
		source.HealthCheckType != "" ||
		source.HealthCheckHTTPEndpoint != "" ||
		source.HealthCheckInvocationTimeout != 0
}

func (source *Application) hasK8sHealthCheckFields() bool {
	return source.StartupProbe != nil ||
		source.LivenessProbe != nil ||
		source.ReadinessProbe != nil
}

// ToContainer converts the manifest to a container suitable for use in a pod,
// ksvc, or app.
func (source *Application) ToContainer(runtimeConfig *v1alpha1.SpaceStatusRuntimeConfig) (corev1.Container, error) {
	resourceRequirements, err := source.ToResourceRequirements(runtimeConfig)
	if err != nil {
		return corev1.Container{}, err
	}

	container := corev1.Container{
		Args:      source.CommandArgs(),
		Command:   source.CommandEntrypoint(),
		Resources: *resourceRequirements,
	}

	if source.hasK8sHealthCheckFields() {
		container.LivenessProbe = source.LivenessProbe
		container.ReadinessProbe = source.ReadinessProbe
		container.StartupProbe = source.StartupProbe
	} else {
		postStartupHealthCheck, err := source.ToPostStartupHealthCheck()
		if err != nil {
			return corev1.Container{}, err
		}
		startupHealthCheck, err := source.ToStartupHealthCheck()
		if err != nil {
			return corev1.Container{}, err
		}

		container.ReadinessProbe = postStartupHealthCheck
		container.LivenessProbe = postStartupHealthCheck
		container.StartupProbe = startupHealthCheck
	}

	if len(source.Env) > 0 {
		container.Env = envutil.MapToEnvVars(source.Env)
	}

	for _, port := range source.Ports {
		container.Ports = append(container.Ports, corev1.ContainerPort{
			Name:          fmt.Sprintf("%s-%d", port.Protocol, port.Port),
			ContainerPort: port.Port,
			// Protocol here is the L4 protocol, not the L7 protocol.
			Protocol: corev1.ProtocolTCP,
		})
	}

	return container, nil
}
