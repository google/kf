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

package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
)

func TestAppSpec_SetDefaults_BlankContainer(t *testing.T) {
	t.Parallel()

	app := &App{}
	app.SetDefaults(context.Background())

	testutil.AssertEqual(t, "len(spec.template.spec.containers)", 1, len(app.Spec.Template.Spec.Containers))
	testutil.AssertEqual(t, "spec.template.spec.containers.name", "", app.Spec.Template.Spec.Containers[0].Name)
}

func TestSetKfAppContainerDefaults(t *testing.T) {
	cases := map[string]struct {
		template *corev1.Container
		expected *corev1.Container
	}{
		"default to TCP": {
			template: &corev1.Container{},
			expected: &corev1.Container{
				ReadinessProbe: &corev1.Probe{
					TimeoutSeconds: DefaultHealthCheckProbeTimeout,
					Handler: corev1.Handler{
						TCPSocket: &corev1.TCPSocketAction{},
					},
				},
			},
		},
		"http path gets defaulted": {
			template: &corev1.Container{
				ReadinessProbe: &corev1.Probe{
					TimeoutSeconds: DefaultHealthCheckProbeTimeout,
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{},
					},
				},
			},
			expected: &corev1.Container{
				ReadinessProbe: &corev1.Probe{
					TimeoutSeconds: DefaultHealthCheckProbeTimeout,
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{Path: DefaultHealthCheckProbeEndpoint},
					},
				},
			},
		},
		"full http doesn't get overwritten": {
			template: &corev1.Container{
				ReadinessProbe: &corev1.Probe{
					TimeoutSeconds: 180,
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"},
					},
				},
			},
			expected: &corev1.Container{
				ReadinessProbe: &corev1.Probe{
					TimeoutSeconds: 180,
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"},
					},
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			SetKfAppContainerDefaults(tc.template)

			testutil.AssertEqual(t, "expected", tc.expected, tc.template)
		})
	}
}
