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
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/apis"
)

func TestValidatePodSpec(t *testing.T) {
	cases := map[string]struct {
		field corev1.PodSpec
		want  *apis.FieldError
	}{
		"empty": {
			field: corev1.PodSpec{},
			want:  apis.ErrMissingField("containers"),
		},
		"invalid field": {
			field: corev1.PodSpec{
				ServiceAccountName: "custom-sa",
				Containers:         []corev1.Container{{}},
			},
			want: apis.ErrDisallowedFields("serviceAccountName"),
		},
		"recurses to containers": {
			field: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						EnvFrom: []corev1.EnvFromSource{{}},
					},
				},
			},
			want: apis.ErrDisallowedFields("containers[0].envFrom"),
		},
		"too many containers": {
			field: corev1.PodSpec{
				Containers: []corev1.Container{{}, {}},
			},
			want: errOnlyOneAllowed("container").ViaField("containers"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := ValidatePodSpec(tc.field)

			testutil.AssertEqual(t, "error", tc.want.Error(), actual.Error())
		})
	}
}

func goodContainer() corev1.Container {
	return corev1.Container{
		Name:       "my-container",
		Command:    []string{"java"},
		Args:       []string{"-jar", "my-app.jar"},
		WorkingDir: "/home/vcap",
		Ports: []corev1.ContainerPort{
			goodContainerPort(),
		},
		Env: []corev1.EnvVar{
			{Name: "FOO", Value: "BAR"},
		},
		Resources:      goodResourceRequirements(),
		LivenessProbe:  goodHTTPContainerProbe(),
		ReadinessProbe: goodTCPContainerProbe(),
	}
}

func TestValidateContainer(t *testing.T) {
	cases := map[string]struct {
		field corev1.Container
		want  *apis.FieldError
	}{
		"good": {
			field: goodContainer(),
			want:  nil,
		},
		"does field mask": {
			field: func() (container corev1.Container) {
				container = goodContainer()
				container.EnvFrom = []corev1.EnvFromSource{{}}

				return
			}(),
			want: apis.ErrDisallowedFields("envFrom"),
		},
		"recurses to ports": {
			field: func() (container corev1.Container) {
				container = goodContainer()
				container.Ports = []corev1.ContainerPort{badContainerPortLower()}

				return
			}(),
			want: apis.ErrOutOfBoundsValue(0, 1, math.MaxUint16, "ports[0].containerPort"),
		},
		"recurses to env": {
			field: func() (container corev1.Container) {
				container = goodContainer()
				container.Env = []corev1.EnvVar{
					{Name: "OK"},
					{Value: "VALUE ONLY1"},
					{Name: "OK2"},
					{Value: "VALUE ONLY2"},
					{Name: "OK3"},
				}

				return
			}(),
			want: apis.ErrMissingField("env[1].name", "env[3].name"),
		},
		"recurses to resources": {
			field: func() (container corev1.Container) {
				container = goodContainer()
				container.Resources = badResourceRequirementsLimits()

				return
			}(),
			want: apis.ErrInvalidKeyName(string(corev1.ResourceServices), "resources.limits"),
		},
		"recurses to liveness probe": {
			field: func() (container corev1.Container) {
				container = goodContainer()
				container.LivenessProbe = badContainerProbe()

				return
			}(),
			want: apis.ErrDisallowedFields("livenessProbe.exec"),
		},
		"recurses to readiness probe": {
			field: func() (container corev1.Container) {
				container = goodContainer()
				container.ReadinessProbe = badContainerProbe()

				return
			}(),
			want: apis.ErrDisallowedFields("readinessProbe.exec"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := ValidateContainer(tc.field)

			testutil.AssertEqual(t, "error", tc.want.Error(), actual.Error())
		})
	}
}

func TestValidateContainerPortsArray(t *testing.T) {
	cases := map[string]struct {
		field []corev1.ContainerPort
		want  *apis.FieldError
	}{
		"empty": {
			field: []corev1.ContainerPort{},
			want:  nil,
		},
		"good": {
			field: []corev1.ContainerPort{
				goodContainerPort(),
			},
			want: nil,
		},
		"bad port": {
			field: []corev1.ContainerPort{
				badContainerPortLower(),
			},
			want: apis.ErrOutOfBoundsValue(0, 1, math.MaxUint16, "[0].containerPort"),
		},
		"duplicate port name": {
			field: []corev1.ContainerPort{
				goodContainerPort(),
				goodContainerPort(),
			},
			want: ErrDuplicateValue("http-8080", "[1].name"),
		},
		"multiple ports": {
			field: []corev1.ContainerPort{
				manifestPort("http2", 8080),
				manifestPort("tcp", 9090),
			},
			want: nil,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := ValidateContainerPortsArray(tc.field)

			testutil.AssertEqual(t, "error", tc.want.Error(), actual.Error())
		})
	}
}

func manifestPort(protocol string, port int32) corev1.ContainerPort {
	return corev1.ContainerPort{
		Name:          fmt.Sprintf("%s-%d", protocol, port),
		ContainerPort: port,
		Protocol:      corev1.ProtocolTCP,
	}
}

func goodContainerPort() corev1.ContainerPort {
	return manifestPort("http", 8080)
}

func badContainerPortLower() corev1.ContainerPort {
	return manifestPort("http", 0)
}

func badContainerPortUpper() corev1.ContainerPort {
	return manifestPort("http", int32(math.MaxUint16)+1)
}

func TestValidateContainerPort(t *testing.T) {
	cases := map[string]struct {
		field corev1.ContainerPort
		want  *apis.FieldError
	}{
		"good": {
			field: goodContainerPort(),
			want:  nil,
		},
		"bad lower-bound": {
			field: badContainerPortLower(),
			want:  apis.ErrOutOfBoundsValue(0, 1, math.MaxUint16, "containerPort"),
		},
		"bad upper-bound": {
			field: badContainerPortUpper(),
			want:  apis.ErrOutOfBoundsValue(int32(math.MaxUint16)+1, 1, math.MaxUint16, "containerPort"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := ValidateContainerPort(tc.field)

			testutil.AssertEqual(t, "error", tc.want.Error(), actual.Error())
		})
	}
}

func TestValidateContainerEnv(t *testing.T) {
	cases := map[string]struct {
		field corev1.EnvVar
		want  *apis.FieldError
	}{
		"empty value is okay": {
			field: corev1.EnvVar{Name: "some-name"},
			want:  nil,
		},
		"fully populated": {
			field: corev1.EnvVar{Name: "some-name", Value: "some-value"},
			want:  nil,
		},
		"no name": {
			field: corev1.EnvVar{Value: "some-value"},
			want:  apis.ErrMissingField("name"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := ValidateContainerEnv(tc.field)

			testutil.AssertEqual(t, "error", tc.want.Error(), actual.Error())
		})
	}
}

func goodResourceRequirements() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:              resource.MustParse("100m"),
			corev1.ResourceEphemeralStorage: resource.MustParse("1G"),
			corev1.ResourceMemory:           resource.MustParse("1G"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:              resource.MustParse("50m"),
			corev1.ResourceEphemeralStorage: resource.MustParse("512M"),
			corev1.ResourceMemory:           resource.MustParse("512M"),
		},
	}
}

func badResourceRequirementsLimits() corev1.ResourceRequirements {
	reqs := goodResourceRequirements()
	reqs.Limits[corev1.ResourceServices] = resource.MustParse("300")
	return reqs
}

func badResourceRequirementsRequests() corev1.ResourceRequirements {
	reqs := goodResourceRequirements()
	reqs.Requests[corev1.ResourceServices] = resource.MustParse("300")
	return reqs
}

func TestValidateContainerResources(t *testing.T) {
	cases := map[string]struct {
		field corev1.ResourceRequirements
		want  *apis.FieldError
	}{
		"good requirements": {
			field: goodResourceRequirements(),
			want:  nil,
		},
		"bad limits": {
			field: badResourceRequirementsLimits(),
			want:  apis.ErrInvalidKeyName(string(corev1.ResourceServices), "limits"),
		},
		"bad requests": {
			field: badResourceRequirementsRequests(),
			want:  apis.ErrInvalidKeyName(string(corev1.ResourceServices), "requests"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := ValidateContainerResources(tc.field)

			testutil.AssertEqual(t, "error", tc.want.Error(), actual.Error())
		})
	}
}

func goodHTTPContainerProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
			},
		},
	}
}

func goodTCPContainerProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{},
		},
	}
}

func badContainerProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{},
		},
	}
}

func TestValidateContainerProbe(t *testing.T) {
	cases := map[string]struct {
		field *corev1.Probe
		want  *apis.FieldError
	}{
		"good HTTP": {
			field: goodHTTPContainerProbe(),
			want:  nil,
		},
		"good TCP": {
			field: goodTCPContainerProbe(),
			want:  nil,
		},
		"multiple probe handlers invalid": {
			field: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{

					HTTPGet: &corev1.HTTPGetAction{
						Path: "/",
					},
					TCPSocket: &corev1.TCPSocketAction{},
				},
			},
			want: apis.ErrMultipleOneOf("tcpSocket", "httpGet"),
		},
		"exec not allowed": {
			field: badContainerProbe(),
			want:  apis.ErrDisallowedFields("exec"),
		},
		"no probe": {
			field: nil,
			want:  nil,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := ValidateContainerProbe(tc.field)

			testutil.AssertEqual(t, "error", tc.want.Error(), actual.Error())
		})
	}
}

func TestValidatePortNumberBounds(t *testing.T) {
	cases := map[string]struct {
		port int32
		want *apis.FieldError
	}{
		"too low": {
			port: 0,
			want: apis.ErrOutOfBoundsValue(0, 1, math.MaxUint16, "fieldname"),
		},
		"good": {
			port: 9000,
			want: nil,
		},
		"too high": {
			port: math.MaxUint16 + 1,
			want: apis.ErrOutOfBoundsValue(math.MaxUint16+1, 1, math.MaxUint16, "fieldname"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := ValidatePortNumberBounds(tc.port, "fieldname")

			testutil.AssertEqual(t, "error", tc.want.Error(), actual.Error())
		})
	}
}
