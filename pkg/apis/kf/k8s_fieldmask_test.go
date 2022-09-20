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
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestPodSpecMask(t *testing.T) {
	want := corev1.PodSpec{
		Containers: []corev1.Container{{
			Image: "helloworld",
		}},
	}
	in := corev1.PodSpec{
		ServiceAccountName: "default",
		Containers: []corev1.Container{{
			Image: "helloworld",
		}},
		Volumes: []corev1.Volume{{
			Name: "the-name",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "foo",
				},
			},
		}},
		InitContainers: []corev1.Container{{
			Image: "busybox",
		}},
	}

	got := podSpecMask(in)

	testutil.AssertEqual(t, "masked value", want, got)
}

func TestContainerMask(t *testing.T) {
	want := corev1.Container{
		Name:           "foo",
		Command:        []string{"world"},
		Args:           []string{"hello"},
		WorkingDir:     "/home/vcap",
		Ports:          []corev1.ContainerPort{{}},
		Env:            []corev1.EnvVar{{}},
		Resources:      corev1.ResourceRequirements{},
		LivenessProbe:  &corev1.Probe{},
		ReadinessProbe: &corev1.Probe{},
		StartupProbe:   &corev1.Probe{},
	}
	in := corev1.Container{
		Name:           "foo",
		Command:        []string{"world"},
		Args:           []string{"hello"},
		WorkingDir:     "/home/vcap",
		Ports:          []corev1.ContainerPort{{}},
		Env:            []corev1.EnvVar{{}},
		Resources:      corev1.ResourceRequirements{},
		LivenessProbe:  &corev1.Probe{},
		ReadinessProbe: &corev1.Probe{},
		StartupProbe:   &corev1.Probe{},

		Image:                    "python",
		EnvFrom:                  []corev1.EnvFromSource{{}},
		ImagePullPolicy:          corev1.PullAlways,
		VolumeMounts:             []corev1.VolumeMount{{}},
		VolumeDevices:            []corev1.VolumeDevice{{}},
		SecurityContext:          &corev1.SecurityContext{},
		TerminationMessagePath:   "/",
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		Lifecycle:                &corev1.Lifecycle{},
		Stdin:                    true,
		StdinOnce:                true,
		TTY:                      true,
	}

	got := containerMask(in)

	testutil.AssertEqual(t, "masked value", want, got)
}

func TestContainerProbeMask(t *testing.T) {
	want := corev1.Probe{
		ProbeHandler:        corev1.ProbeHandler{},
		InitialDelaySeconds: 42,
		TimeoutSeconds:      42,
		PeriodSeconds:       42,
		SuccessThreshold:    42,
		FailureThreshold:    42,
	}
	in := want

	got := containerProbeMask(in)

	testutil.AssertEqual(t, "masked value", want, got)
}

func TestContainerProbeHandlerMask(t *testing.T) {
	want := corev1.ProbeHandler{
		HTTPGet:   &corev1.HTTPGetAction{},
		TCPSocket: &corev1.TCPSocketAction{},
	}
	in := corev1.ProbeHandler{
		Exec:      &corev1.ExecAction{},
		HTTPGet:   &corev1.HTTPGetAction{},
		TCPSocket: &corev1.TCPSocketAction{},
	}

	got := containerProbeHandlerMask(in)

	testutil.AssertEqual(t, "masked value", want, got)
}

func TestContainerProbeHandlerHTTPGetActionMask(t *testing.T) {
	want := corev1.HTTPGetAction{
		Host:        "foo",
		Path:        "/bar",
		Scheme:      corev1.URISchemeHTTP,
		HTTPHeaders: []corev1.HTTPHeader{{}},
	}
	in := corev1.HTTPGetAction{
		Host:        "foo",
		Path:        "/bar",
		Scheme:      corev1.URISchemeHTTP,
		HTTPHeaders: []corev1.HTTPHeader{{}},
		Port:        intstr.FromInt(8080),
	}

	got := containerProbeHandlerHTTPGetActionMask(in)
	testutil.AssertEqual(t, "masked value", want, got)
}

func TestContainerProbeHandlerTCPSocketActionMask(t *testing.T) {
	want := corev1.TCPSocketAction{
		Host: "foo",
	}
	in := corev1.TCPSocketAction{
		Host: "foo",
		Port: intstr.FromInt(8080),
	}

	got := containerProbeHandlerTCPSocketActionMask(in)
	testutil.AssertEqual(t, "masked value", want, got)
}

func TestContainerPortMask(t *testing.T) {
	want := corev1.ContainerPort{
		ContainerPort: 42,
		Name:          "foo",
		Protocol:      corev1.ProtocolTCP,
	}
	in := corev1.ContainerPort{
		ContainerPort: 42,
		Name:          "foo",
		Protocol:      corev1.ProtocolTCP,
		HostIP:        "10.0.0.1",
		HostPort:      43,
	}

	got := containerPortMask(in)

	testutil.AssertEqual(t, "masked value", want, got)
}

func TestContainerEnvMask(t *testing.T) {
	want := corev1.EnvVar{
		Name:  "foo",
		Value: "bar",
	}
	in := corev1.EnvVar{
		Name:      "foo",
		Value:     "bar",
		ValueFrom: &corev1.EnvVarSource{},
	}

	got := containerEnvMask(in)

	testutil.AssertEqual(t, "masked value", want, got)
}

func TestContainerResourceRequirementsMask(t *testing.T) {
	want := corev1.ResourceRequirements{
		Limits:   make(corev1.ResourceList),
		Requests: make(corev1.ResourceList),
	}
	in := want

	got := containerResourceRequirementsMask(in)

	testutil.AssertEqual(t, "masked value", want, got)
}
