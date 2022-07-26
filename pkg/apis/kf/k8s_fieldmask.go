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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// podSpecMask creates a shallow copy of the input with all unsettable top-level
// fields masked off.
func podSpecMask(in corev1.PodSpec) (out corev1.PodSpec) {
	// Allowed fields
	out.Containers = in.Containers

	// Disallowed fields, too many to list

	return out
}

// containerMask creates a shallow copy of the input with all unsettable top-level
// fields masked off.
func containerMask(in corev1.Container) (out corev1.Container) {

	// Allowed fields (order based on corev1.Container definition)
	out.Name = in.Name
	out.Command = in.Command
	out.Args = in.Args
	out.WorkingDir = in.WorkingDir
	out.Ports = in.Ports
	out.Env = in.Env
	out.Resources = in.Resources
	out.LivenessProbe = in.LivenessProbe
	out.ReadinessProbe = in.ReadinessProbe

	// Explicitly disallowed fields.
	// These are optional, but provided here for clarity.

	out.Image = ""                    // Only allowed from a source
	out.EnvFrom = nil                 // Not supported by the Kf manifest
	out.ImagePullPolicy = ""          // Overridden by Kf platform
	out.VolumeMounts = nil            // Must be specified via service
	out.VolumeDevices = nil           // Must be specified via service
	out.SecurityContext = nil         // Overridden by Kf platform
	out.TerminationMessagePath = ""   // Overridden by Kf platform
	out.TerminationMessagePolicy = "" // Overridden by Kf platform
	out.Lifecycle = nil               // Not supported by the Kf manifest

	// Interactive compnents not supported by the platform:
	out.Stdin = false
	out.StdinOnce = false
	out.TTY = false

	return out
}

// containerEnvMask creates a shallow copy of the input with all unsettable
// top-level fields masked off.
func containerPortMask(in corev1.ContainerPort) (out corev1.ContainerPort) {

	// Allowed fields
	out.ContainerPort = in.ContainerPort
	out.Name = in.Name
	out.Protocol = in.Protocol

	// Disallowed fields
	// This list is unnecessary, but added here for clarity
	out.HostIP = ""  // Kf doesn't allow binding to specific IP addresses.
	out.HostPort = 0 // Automatically set by Kf to work with the routing layer.

	return out
}

// containerEnvMask creates a shallow copy of the input with all unsettable
// top-level fields masked off.
func containerEnvMask(in corev1.EnvVar) (out corev1.EnvVar) {

	// Allowed fields
	out.Name = in.Name
	out.Value = in.Value

	// Disallowed fields

	// ValueFrom is not supported by the Kf manifest, reconciliation with this
	// field could get become tricky to do correctly without explicitly adding
	// support.
	out.ValueFrom = nil

	return out
}

// containerResourceRequirementsMask creates a shallow copy of the input with all
// unsettable top-level fields masked off.
func containerResourceRequirementsMask(in corev1.ResourceRequirements) (out corev1.ResourceRequirements) {

	// Allowed fields
	out.Limits = in.Limits
	out.Requests = in.Requests

	// Disallowed fields
	// As of Kubernetes 1.17 there are no disallowed fields.

	return out
}

// containerProbeMask creates a shallow copy of the input with all unsettable top-level
// fields masked off.
func containerProbeMask(in corev1.Probe) (out corev1.Probe) {
	// Allowed fields
	out.ProbeHandler = in.ProbeHandler
	out.InitialDelaySeconds = in.InitialDelaySeconds
	out.TimeoutSeconds = in.TimeoutSeconds
	out.PeriodSeconds = in.PeriodSeconds
	out.SuccessThreshold = in.SuccessThreshold
	out.FailureThreshold = in.FailureThreshold

	// Disallowed fields
	// As of Kubernetes 1.17 there are no disallowed fields.

	return out
}

// containerProbeHandlerMask creates a shallow copy of the input with all unsettable
// top-level fields masked off.
func containerProbeHandlerMask(in corev1.ProbeHandler) (out corev1.ProbeHandler) {

	// Allowed fields
	out.HTTPGet = in.HTTPGet
	out.TCPSocket = in.TCPSocket

	// Disallowed fields
	out.Exec = nil // Not supported by the Kf manifest

	return out
}

// containerProbeHandlerHTTPGetActionMask creates a shallow copy of the input with all
// unsettable top-level fields masked off.
func containerProbeHandlerHTTPGetActionMask(in corev1.HTTPGetAction) (out corev1.HTTPGetAction) {

	// Allowed fields
	out.Host = in.Host
	out.Path = in.Path
	out.Scheme = in.Scheme
	out.HTTPHeaders = in.HTTPHeaders

	// Disallowed fields
	out.Port = intstr.IntOrString{} // Populated by Kf automatically

	return out
}

// containerProbeHandlerTCPSocketActionMask creates a shallow copy of the input with all
// unsettable top-level fields masked off.
func containerProbeHandlerTCPSocketActionMask(in corev1.TCPSocketAction) (out corev1.TCPSocketAction) {
	// Allowed fields
	out.Host = in.Host

	// Disallowed fields
	out.Port = intstr.IntOrString{} // Populated by Kf automatically

	return out
}
