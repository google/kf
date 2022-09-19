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
	"context"
	"math"
	"testing"

	kfapis "github.com/google/kf/v2/pkg/apis/kf"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/apis"
)

func TestApplication_Validation(t *testing.T) {
	cases := map[string]struct {
		spec Application
		want *apis.FieldError
	}{
		"valid": {
			spec: Application{},
		},
		"entrypoint and args": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					Args:       []string{"-m", "SimpleHTTPServer"},
					Entrypoint: "python",
				},
			},
		},
		"command and args": {
			spec: Application{
				Command: "python",
				KfApplicationExtension: KfApplicationExtension{
					Args: []string{"-m", "SimpleHTTPServer"},
				},
			},
			want: apis.ErrMultipleOneOf("args", "command"),
		},
		"entrypoint and command": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					Entrypoint: "/lifecycle/launcher",
				},
				Command: "python",
			},
			want: apis.ErrMultipleOneOf("entrypoint", "command"),
		},
		"buildpack and buildpacks": {
			spec: Application{
				LegacyBuildpack: "default",
				Buildpacks:      []string{"java", "node"},
			},
			want: apis.ErrMultipleOneOf("buildpack", "buildpacks"),
		},
		"good ports and routes": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					Ports: AppPortList{
						{Port: 8080, Protocol: protocolHTTP},
						{Port: 8081, Protocol: protocolHTTP2},
						{Port: 8082, Protocol: protocolTCP},
					},
				},
				Routes: []Route{
					{Route: "default"},
					{Route: "explicit", AppPort: 8080},
				},
			},
			want: nil,
		},
		"duplicate port": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					Ports: AppPortList{
						{Port: 8080, Protocol: protocolHTTP},
						{Port: 8080, Protocol: protocolHTTP2},
					},
				},
			},
			want: kfapis.ErrDuplicateValue(8080, "ports[1].port"),
		},
		"bad protocol": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					Ports: AppPortList{
						{Port: 8080, Protocol: "foo"},
					},
				},
			},
			want: apis.ErrInvalidValue("must be one of: [http http2 tcp]", "ports[0].protocol"),
		},
		"bad port": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					Ports: AppPortList{
						{Port: 80808080, Protocol: "tcp"},
					},
				},
			},
			want: apis.ErrOutOfBoundsValue(80808080, 1, math.MaxUint16, "ports[0].port"),
		},
		"route port missing": {
			spec: Application{
				Routes: []Route{
					{Route: "missing-port", AppPort: 8080},
				},
			},
			want: apis.ErrInvalidValue("must match a declared port", "routes[0].appPort"),
		},
		"valid probes": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					StartupProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/warmed-up",
							},
						},
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/healthz",
							},
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{},
						},
					},
				},
			},
		},
		"grpc disabled": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					StartupProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							GRPC: &corev1.GRPCAction{},
						},
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							GRPC: &corev1.GRPCAction{},
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							GRPC: &corev1.GRPCAction{},
						},
					},
				},
			},
			want: apis.ErrDisallowedFields("livenessProbe.grpc", "readinessProbe.grpc", "startupProbe.grpc"),
		},
		"exec disabled": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					StartupProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{},
						},
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{},
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{},
						},
					},
				},
			},
			want: apis.ErrDisallowedFields("livenessProbe.exec", "readinessProbe.exec", "startupProbe.exec"),
		},
		"http port disallowed": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					StartupProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Port: intstr.FromInt(9999),
							},
						},
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Port: intstr.FromInt(9999),
							},
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Port: intstr.FromInt(9999),
							},
						},
					},
				},
			},
			want: apis.ErrDisallowedFields("livenessProbe.httpGet.port", "readinessProbe.httpGet.port", "startupProbe.httpGet.port"),
		},
		"tcp port disallowed": {
			spec: Application{
				KfApplicationExtension: KfApplicationExtension{
					StartupProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.FromInt(9999),
							},
						},
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.FromInt(9999),
							},
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.FromInt(9999),
							},
						},
					},
				},
			},
			want: apis.ErrDisallowedFields("livenessProbe.tcpSocket.port", "readinessProbe.tcpSocket.port", "startupProbe.tcpSocket.port"),
		},
		"conflicting startupProbe": {
			spec: Application{
				HealthCheckTimeout: 300,
				KfApplicationExtension: KfApplicationExtension{
					StartupProbe: &corev1.Probe{},
				},
			},
			want: &apis.FieldError{Message: "startupProbe, livenessProbe, and readinessProbe can't be used with CF health check fields"},
		},
		"conflicting livenessProbe": {
			spec: Application{
				HealthCheckType: "http",
				KfApplicationExtension: KfApplicationExtension{
					LivenessProbe: &corev1.Probe{},
				},
			},
			want: &apis.FieldError{Message: "startupProbe, livenessProbe, and readinessProbe can't be used with CF health check fields"},
		},
		"conflicting readinessProbe": {
			spec: Application{
				HealthCheckHTTPEndpoint: "/",
				KfApplicationExtension: KfApplicationExtension{
					ReadinessProbe: &corev1.Probe{},
				},
			},
			want: &apis.FieldError{Message: "startupProbe, livenessProbe, and readinessProbe can't be used with CF health check fields"},
		},
		"good http health check": {
			spec: Application{
				HealthCheckTimeout:           30,
				HealthCheckInvocationTimeout: 30,
				HealthCheckType:              "http",
				HealthCheckHTTPEndpoint:      "/statusz",
			},
		},
		"bad health check type": {
			spec: Application{
				HealthCheckType: "foo",
			},
			want: apis.ErrInvalidValue("foo", "health-check-type", `valid values are: ["" "http" "none" "port" "process"]`),
		},
		"http endpoint is only valid with http": {
			spec: Application{
				HealthCheckType:         "port",
				HealthCheckHTTPEndpoint: "/",
			},
			want: apis.ErrInvalidValue("/", "health-check-http-endpoint", `field can only be set if health-check-type is "http"`),
		},
		"good port health check": {
			spec: Application{
				HealthCheckTimeout:           30,
				HealthCheckInvocationTimeout: 30,
				HealthCheckType:              "port",
			},
		},
		"good process health check": {
			spec: Application{
				HealthCheckType: "process",
			},
		},
		"good none (process) health check": {
			spec: Application{
				HealthCheckType: "none",
			},
		},
		"bad health check timeout": {
			spec: Application{
				HealthCheckTimeout: -1,
			},
			want: apis.ErrInvalidValue(-1, "timeout", "health check timeout can't be negative"),
		},
		"bad health check invocation timeout": {
			spec: Application{
				HealthCheckInvocationTimeout: -2,
			},
			want: apis.ErrInvalidValue(-2, "health-check-invocation-timeout", "health check timeout can't be negative"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}

func TestApplicationMetadata_Validation(t *testing.T) {
	cases := map[string]struct {
		spec ApplicationMetadata
		want *apis.FieldError
	}{
		// Labels and annotations rely on K8s validation, but we still want to assert
		// the formats and paths look okay.

		"valid": {
			spec: ApplicationMetadata{
				Labels: map[string]string{
					"kf.dev/test": "ok",
				},
				Annotations: map[string]string{
					"kf.dev/test": "ok",
				},
			},
			want: nil,
		},
		"blank okay": {
			spec: ApplicationMetadata{},
			want: nil,
		},

		"bad label key": {
			spec: ApplicationMetadata{
				Labels: map[string]string{
					"kf.dev test": "bad",
				},
			},
			want: &apis.FieldError{
				Message: `Invalid value: "kf.dev test": name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')`,
				Paths:   []string{"labels"},
			},
		},
		"bad annotation key": {
			spec: ApplicationMetadata{
				Annotations: map[string]string{
					"kf.dev test": "bad",
				},
			},
			want: &apis.FieldError{
				Message: `Invalid value: "kf.dev test": name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')`,
				Paths:   []string{"annotations"},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := tc.spec.Validate(context.Background())

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}

}
