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

package kf

import (
	"fmt"
	"math"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
)

var (
	// allowedResourceRequirements contains the subset of resource requirements we
	// support on containers
	allowedResourceRequirements = sets.NewString(
		string(corev1.ResourceCPU),
		string(corev1.ResourceMemory),
		string(corev1.ResourceEphemeralStorage),
	)
)

func errOnlyOneAllowed(fieldType string) *apis.FieldError {
	return &apis.FieldError{
		Message: fmt.Sprintf("More than one %s is set", fieldType),
		Paths:   []string{apis.CurrentField},
		Details: fmt.Sprintf("Only a single %s is allowed", fieldType),
	}
}

// ErrDuplicateValue creates a FieldError that indicates a value must not be
// duplicated.
func ErrDuplicateValue(value interface{}, fieldPaths ...string) *apis.FieldError {
	return &apis.FieldError{
		Message: fmt.Sprintf("Duplicate value: %v", value),
		Paths:   fieldPaths,
	}
}

// ValidatePodSpec performs a deep validation that the PodSpec matches Kf's
// expectations.
func ValidatePodSpec(podSpec corev1.PodSpec) *apis.FieldError {
	errs := apis.CheckDisallowedFields(podSpec, podSpecMask(podSpec))

	switch len(podSpec.Containers) {
	case 0:
		errs = errs.Also(apis.ErrMissingField("containers"))
	case 1:
		errs = errs.Also(ValidateContainer(podSpec.Containers[0]).ViaFieldIndex("containers", 0))
	default:
		errs = errs.Also(errOnlyOneAllowed("container").ViaField("containers"))
	}
	return errs
}

// ValidateContainer performs a deep validation that the Container matches Kf's
// expectations.
func ValidateContainer(container corev1.Container) *apis.FieldError {
	errs := apis.CheckDisallowedFields(container, containerMask(container))

	// Skip validating: Name, Command, Args, WorkingDir
	// because they need additional container info to validate.

	errs = errs.Also(ValidateContainerPortsArray(container.Ports).ViaField("ports"))

	for idx, env := range container.Env {
		errs = errs.Also(ValidateContainerEnv(env).ViaFieldIndex("env", idx))
	}

	errs = errs.Also(ValidateContainerResources(container.Resources).ViaField("resources"))
	errs = errs.Also(ValidateContainerProbe(container.LivenessProbe).ViaField("livenessProbe"))
	errs = errs.Also(ValidateContainerProbe(container.ReadinessProbe).ViaField("readinessProbe"))
	errs = errs.Also(ValidateContainerProbe(container.StartupProbe).ViaField("startupProbe"))

	return errs
}

// ValidateContainerPortsArray performs a deep validation that the Port matches
// Kf's expectations.
func ValidateContainerPortsArray(ports []corev1.ContainerPort) (errs *apis.FieldError) {
	if len(ports) == 0 {
		return nil
	}

	seenNames := sets.NewString()
	for idx, port := range ports {
		errs = errs.Also(ValidateContainerPort(port).ViaIndex(idx))

		// names can't be duplicated, but port numbers can in Kubernetes
		if port.Name != "" && seenNames.Has(port.Name) {
			errs = errs.Also(ErrDuplicateValue(port.Name, "name").ViaIndex(idx))
		}

		seenNames.Insert(port.Name)
	}

	return errs
}

// ValidateContainerPort validates a specific port.
func ValidateContainerPort(port corev1.ContainerPort) *apis.FieldError {
	errs := apis.CheckDisallowedFields(port, containerPortMask(port))
	errs = errs.Also(ValidatePortNumberBounds(port.ContainerPort, "containerPort"))

	return errs
}

// ValidatePortNumberBounds checks that the given port number is within
// acceptable ranges for a port.
func ValidatePortNumberBounds(port int32, fieldName string) *apis.FieldError {
	if port < 1 || port > math.MaxUint16 {
		return apis.ErrOutOfBoundsValue(port, 1, math.MaxUint16, fieldName)
	}

	return nil
}

// ValidateContainerEnv performs a deep validation that the Environment Variable
// matches Kf's expectations.
func ValidateContainerEnv(env corev1.EnvVar) *apis.FieldError {
	errs := apis.CheckDisallowedFields(env, containerEnvMask(env))

	if env.Name == "" {
		errs = errs.Also(apis.ErrMissingField("name"))
	}

	// FIXME(jlewisiii) once b/145938633 is done keys should be checked against
	// those we set in the environment vartiable secret and be disallowed.
	// Alternatively, only non-overridden values could be set in that secret.

	return errs
}

// ValidateContainerResources performs a deep validation that the quota
// matches Kf's expectations.
func ValidateContainerResources(requirements corev1.ResourceRequirements) *apis.FieldError {
	errs := apis.CheckDisallowedFields(requirements, containerResourceRequirementsMask(requirements))

	for k := range requirements.Limits {
		keyName := string(k)

		if !allowedResourceRequirements.Has(keyName) {
			errs = errs.Also(apis.ErrInvalidKeyName(keyName, "limits"))
		}
	}

	for k := range requirements.Requests {
		keyName := string(k)

		if !allowedResourceRequirements.Has(keyName) {
			errs = errs.Also(apis.ErrInvalidKeyName(keyName, "requests"))
		}
	}

	return errs
}

// ValidateContainerProbe performs a deep validation that the probe
// matches Kf's expectations.
func ValidateContainerProbe(probe *corev1.Probe) *apis.FieldError {
	if probe == nil {
		return nil
	}

	errs := apis.CheckDisallowedFields(*probe, containerProbeMask(*probe))

	handler := probe.ProbeHandler
	errs = errs.Also(apis.CheckDisallowedFields(handler, containerProbeHandlerMask(handler)))

	// NOTE: the checks here should match the probes in containerProbeHandlerMask
	suppliedHandlers := sets.NewString()
	if handler.HTTPGet != nil {
		suppliedHandlers.Insert("httpGet")
	}
	if handler.TCPSocket != nil {
		suppliedHandlers.Insert("tcpSocket")
	}
	switch {
	case suppliedHandlers.Len() > 1:
		errs = errs.Also(apis.ErrMultipleOneOf(suppliedHandlers.List()...))
	case handler.HTTPGet != nil:
		masked := containerProbeHandlerHTTPGetActionMask(*handler.HTTPGet)
		errs = errs.Also(apis.CheckDisallowedFields(*handler.HTTPGet, masked)).ViaField("httpGet")
	case handler.TCPSocket != nil:
		masked := containerProbeHandlerTCPSocketActionMask(*handler.TCPSocket)
		errs = errs.Also(apis.CheckDisallowedFields(*handler.TCPSocket, masked)).ViaField("tcpSocket")
	}

	return errs
}
