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

package apps

import (
	"errors"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
)

func TestNewHealthCheck(t *testing.T) {
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
		"http complete": {
			checkType: "http",
			endpoint:  "/healthz",
			timeout:   180,
			expectProbe: &corev1.Probe{
				TimeoutSeconds: int32(180),
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
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{},
				},
			},
		},
		"blank type uses port": {
			expectProbe: &corev1.Probe{
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
				TimeoutSeconds: int32(180),
				Handler: corev1.Handler{
					TCPSocket: &corev1.TCPSocketAction{},
				},
			},
		},
		"port default": {
			checkType: "port",
			expectProbe: &corev1.Probe{
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
			actualProbe, actualErr := NewHealthCheck(tc.checkType, tc.endpoint, tc.timeout)

			testutil.AssertErrorsEqual(t, tc.expectErr, actualErr)
			testutil.AssertEqual(t, "probe", tc.expectProbe, actualProbe)
		})
	}
}
