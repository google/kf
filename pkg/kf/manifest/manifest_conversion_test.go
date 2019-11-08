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

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
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
		"stopped autoscaled app": {
			source: Application{
				KfApplicationExtension: KfApplicationExtension{
					NoStart:  ptr.Bool(true),
					MinScale: intPtr(2),
					MaxScale: intPtr(300),
				},
			},
			expected: v1alpha1.AppSpecInstances{
				Stopped: true,
				Min:     intPtr(2),
				Max:     intPtr(300),
			},
		},
		"started app with instances": {
			source: Application{
				Instances: intPtr(3),
				KfApplicationExtension: KfApplicationExtension{
					NoStart: ptr.Bool(false),
				},
			},
			expected: v1alpha1.AppSpecInstances{
				Exactly: intPtr(3),
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
			checkType: "process",
			expectErr: errors.New("kf doesn't support the process health check type"),
		},
		"none is process type": {
			checkType: "none",
			expectErr: errors.New("kf doesn't support the process health check type"),
		},
		"http complete": {
			checkType: "http",
			endpoint:  "/healthz",
			timeout:   180,
			expectProbe: &corev1.Probe{
				TimeoutSeconds:   int32(180),
				SuccessThreshold: 1,
				Handler: corev1.Handler{
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
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
			},
		},
		"blank type uses port": {
			expectProbe: &corev1.Probe{
				SuccessThreshold: 1,
				Handler: corev1.Handler{
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
				Handler: corev1.Handler{
					TCPSocket: &corev1.TCPSocketAction{},
				},
			},
		},
		"port default": {
			checkType: "port",
			expectProbe: &corev1.Probe{
				SuccessThreshold: 1,
				Handler: corev1.Handler{
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
		Handler: corev1.Handler{
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
					Args:        []string{"foo", "bar"},
					Entrypoint:  "bash",
					EnableHTTP2: ptr.Bool(true),
				},
			},
			expectContainer: corev1.Container{
				Args:    []string{"foo", "bar"},
				Command: []string{"bash"},
				ReadinessProbe: &corev1.Probe{
					SuccessThreshold: 1,
					Handler: corev1.Handler{
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
				Ports: HTTP2ContainerPort(),
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
