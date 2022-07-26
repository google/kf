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
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/ptr"
)

func TestApplication_ToResourceRequests(t *testing.T) {
	cases := map[string]struct {
		source       Application
		expectedList corev1.ResourceList
		expectedErr  error
	}{
		"full": {
			source: Application{
				Memory:    "30MB", // CF uses X and XB as SI units, these get changed to SI
				DiskQuota: "1G",
				KfApplicationExtension: KfApplicationExtension{
					CPU: "200m",
				},
			},
			expectedList: corev1.ResourceList{
				corev1.ResourceMemory:           resource.MustParse("30Mi"),
				corev1.ResourceCPU:              resource.MustParse("200m"),
				corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
			},
		},
		"normal cf subset": {
			source: Application{
				Memory:    "30M",
				DiskQuota: "1Gi",
			},
			expectedList: corev1.ResourceList{
				corev1.ResourceMemory:           resource.MustParse("30Mi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
			},
		},
		"bad quantity": {
			source: Application{
				Memory: "30Y",
			},
			expectedErr: errors.New("couldn't parse resource quantity 30Y: quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'"),
		},
		"no quotas": {
			source:       Application{},
			expectedList: nil, // explicitly want nil rather than the empty map
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualList, actualErr := tc.source.ToResourceRequests()

			testutil.AssertErrorsEqual(t, tc.expectedErr, actualErr)
			testutil.AssertEqual(t, "resource lists", tc.expectedList, actualList)
		})
	}
}

func TestApplication_ToAppSpecInstances(t *testing.T) {
	cases := map[string]struct {
		source   Application
		expected v1alpha1.AppSpecInstances
	}{
		"blank app": {
			source:   Application{},
			expected: v1alpha1.AppSpecInstances{},
		},
		"stopped app": {
			source: Application{
				Instances: ptr.Int32(5),
				KfApplicationExtension: KfApplicationExtension{
					NoStart: ptr.Bool(true),
				},
			},
			expected: v1alpha1.AppSpecInstances{
				Stopped:  true,
				Replicas: ptr.Int32(5),
			},
		},
		"app for task": {
			source: Application{
				Task: ptr.Bool(true),
			},
			expected: v1alpha1.AppSpecInstances{
				Stopped: true,
			},
		},
		"started app with instances": {
			source: Application{
				Instances: ptr.Int32(3),
				KfApplicationExtension: KfApplicationExtension{
					NoStart: ptr.Bool(false),
				},
			},
			expected: v1alpha1.AppSpecInstances{
				Replicas: ptr.Int32(3),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := tc.source.ToAppSpecInstances()

			testutil.AssertEqual(t, "instances", tc.expected, actual)
		})
	}
}

func TestApplication_ToHealthCheck(t *testing.T) {
	cases := map[string]struct {
		checkType string
		endpoint  string
		timeout   int

		expectProbe *corev1.Probe
		expectErr   error
	}{
		"invalid type": {
			checkType: "foo",
			expectErr: errors.New("unknown health check type foo, supported types are http and port"),
		},
		"process type": {
			checkType:   "process",
			expectProbe: nil,
		},
		"none is process type": {
			checkType:   "none",
			expectProbe: nil,
		},
		"http complete": {
			checkType: "http",
			endpoint:  "/healthz",
			timeout:   180,
			expectProbe: &corev1.Probe{
				TimeoutSeconds:   int32(180),
				SuccessThreshold: 1,
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/healthz",
					},
				},
			},
		},
		"http default": {
			checkType: "http",
			expectProbe: &corev1.Probe{
				SuccessThreshold: 1,
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
			},
		},
		"blank type uses port": {
			expectProbe: &corev1.Probe{
				SuccessThreshold: 1,
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{},
				},
			},
		},
		"negative timeout": {
			timeout:   -1,
			expectErr: errors.New("health check timeouts can't be negative"),
		},
		"port complete": {
			checkType: "port",
			timeout:   180,
			expectProbe: &corev1.Probe{
				TimeoutSeconds:   int32(180),
				SuccessThreshold: 1,
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{},
				},
			},
		},
		"port default": {
			checkType: "port",
			expectProbe: &corev1.Probe{
				SuccessThreshold: 1,
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{},
				},
			},
		},
		"port with endpoint": {
			checkType: "port",
			endpoint:  "/healthz",
			expectErr: errors.New("health check endpoints can only be used with http checks"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			app := Application{
				HealthCheckType:         tc.checkType,
				HealthCheckHTTPEndpoint: tc.endpoint,
				HealthCheckTimeout:      tc.timeout,
			}

			actualProbe, actualErr := app.ToHealthCheck()

			testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
			testutil.AssertEqual(t, "probe", tc.expectProbe, actualProbe)
		})
	}
}

func TestApplication_ToContainer(t *testing.T) {
	defaultHealthCheck := &corev1.Probe{
		SuccessThreshold: 1,
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{},
		},
	}

	cases := map[string]struct {
		app             Application
		expectContainer corev1.Container
		expectErr       error
	}{
		"empty manifest": {
			app: Application{},
			expectContainer: corev1.Container{
				ReadinessProbe: defaultHealthCheck,
			},
		},
		"bad resource requests": {
			app: Application{
				Memory: "21ZB",
			},
			expectErr: errors.New("couldn't parse resource quantity 21ZB: quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'"),
		},
		"bad health check": {
			app: Application{
				HealthCheckType: "NOT ALLOWED",
			},
			expectErr: errors.New("unknown health check type NOT ALLOWED, supported types are http and port"),
		},
		"full manifest": {
			app: Application{
				HealthCheckType: "http",
				Memory:          "30M",
				DiskQuota:       "1Gi",
				Env:             map[string]string{"KEYMASTER": "GATEKEEPER"},
				KfApplicationExtension: KfApplicationExtension{
					Args:       []string{"foo", "bar"},
					Entrypoint: "bash",
					Ports: AppPortList{
						{Port: 9000, Protocol: protocolHTTP2},
						{Port: 2525, Protocol: protocolTCP},
					},
				},
			},
			expectContainer: corev1.Container{
				Args:    []string{"foo", "bar"},
				Command: []string{"bash"},
				ReadinessProbe: &corev1.Probe{
					SuccessThreshold: 1,
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{},
					},
				},
				Env: []corev1.EnvVar{
					{Name: "KEYMASTER", Value: "GATEKEEPER"},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory:           resource.MustParse("30Mi"),
						corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
					},
				},
				Ports: []corev1.ContainerPort{
					{Name: "http2-9000", ContainerPort: 9000, Protocol: "TCP"},
					{Name: "tcp-2525", ContainerPort: 2525, Protocol: "TCP"},
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualContainer, actualErr := tc.app.ToContainer()

			testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
			testutil.AssertEqual(t, "container", tc.expectContainer, actualContainer)
		})
	}
}

func TestCFToSIUnits(t *testing.T) {
	cases := map[string]struct {
		input        string
		expectOutput string
	}{
		"T to Ti": {
			input:        "1T",
			expectOutput: "1Ti",
		},
		"G to Gi": {
			input:        "1G",
			expectOutput: "1Gi",
		},
		"M to Mi": {
			input:        "1M",
			expectOutput: "1Mi",
		},
		"K to Ki": {
			input:        "1K",
			expectOutput: "1Ki",
		},
		"Ti is unchanged": {
			input:        "1Ti",
			expectOutput: "1Ti",
		},
		"Gi is unchanged": {
			input:        "1Gi",
			expectOutput: "1Gi",
		},
		"Mi is unchanged": {
			input:        "1Mi",
			expectOutput: "1Mi",
		},
		"Ki is unchanged": {
			input:        "1Ki",
			expectOutput: "1Ki",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actualOutput := CFToSIUnits(tc.input)
			testutil.AssertEqual(t, "conversion", tc.expectOutput, actualOutput)
		})
	}
}
