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

package resources

import (
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/ptr"
)

func TestBuildRuntimeEnvVars(t *testing.T) {
	app := &v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-app",
			Namespace: "my-ns",
			UID:       "12345",
		},
		Spec: v1alpha1.AppSpec{
			Template: v1alpha1.AppSpecTemplate{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: buildContainerPorts(9999),
						},
					},
				},
			},
		},
		Status: v1alpha1.AppStatus{
			Routes: []v1alpha1.AppRouteStatus{{
				URL: "my-app.example.com",
			}},
		},
	}

	cases := map[string]struct {
		app     *v1alpha1.App
		runtime EnvRuntime
		wantOut []corev1.EnvVar
	}{
		"running app": {
			app:     app,
			runtime: CFRunning,
			wantOut: []corev1.EnvVar{
				{Name: "PORT", Value: "9999"},
				{Name: "VCAP_APP_PORT", Value: "$(PORT)"},
				{Name: "CF_INSTANCE_IP", ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				}},
				{Name: "CF_INSTANCE_INTERNAL_IP", Value: "$(CF_INSTANCE_IP)"},
				{Name: "VCAP_APP_HOST", Value: "$(CF_INSTANCE_IP)"},
				{Name: "CF_INSTANCE_PORT", Value: "9999"},
				{Name: "CF_INSTANCE_ADDR", Value: "$(CF_INSTANCE_IP):$(CF_INSTANCE_PORT)"},
				{Name: "CF_INSTANCE_GUID", ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.uid",
					},
				}},
				{Name: "INSTANCE_GUID", Value: "$(CF_INSTANCE_GUID)"},
				{Name: "CF_INSTANCE_INDEX", Value: "0"},
				{Name: "INSTANCE_INDEX", Value: "$(CF_INSTANCE_INDEX)"},
				{Name: "MEMORY_LIMIT", ValueFrom: &corev1.EnvVarSource{
					ResourceFieldRef: &corev1.ResourceFieldSelector{
						Divisor:  memoryDivisor,
						Resource: "limits.memory",
					},
				}},
				{Name: "DISK_LIMIT", ValueFrom: &corev1.EnvVarSource{
					ResourceFieldRef: &corev1.ResourceFieldSelector{
						Divisor:  diskDivisor,
						Resource: "limits.ephemeral-storage",
					},
				}},
				{Name: "LANG", Value: "en_US.UTF-8"},
				// json.Marshal writes values in sorted key order
				{Name: "VCAP_APPLICATION", Value: `{"application_id":"12345","application_name":"my-app","application_uris":["my-app.example.com"],"limits":{"disk":$(DISK_LIMIT),"mem":$(MEMORY_LIMIT)},"name":"my-app","process_id":"12345","process_type":"web","space_name":"my-ns","uris":["my-app.example.com"]}`},
				{Name: "VCAP_SERVICES", ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "VCAP_SERVICES",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "kf-injected-envs-my-app",
						},
						Optional: ptr.Bool(false),
					},
				}},
				{Name: "DATABASE_URL", ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "DATABASE_URL",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "kf-injected-envs-my-app",
						},
						Optional: ptr.Bool(true),
					},
				}},
			},
		},
		"staging app": {
			app:     app,
			runtime: CFStaging,
			wantOut: []corev1.EnvVar{
				{Name: "CF_INSTANCE_IP", ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				}},
				{Name: "CF_INSTANCE_INTERNAL_IP", Value: "$(CF_INSTANCE_IP)"},
				{Name: "VCAP_APP_HOST", Value: "$(CF_INSTANCE_IP)"},
				{Name: "CF_INSTANCE_PORT", Value: "9999"},
				{Name: "CF_INSTANCE_ADDR", Value: "$(CF_INSTANCE_IP):$(CF_INSTANCE_PORT)"},
				{Name: "MEMORY_LIMIT", ValueFrom: &corev1.EnvVarSource{
					ResourceFieldRef: &corev1.ResourceFieldSelector{
						Divisor:  memoryDivisor,
						Resource: "limits.memory",
					},
				}},
				{Name: "DISK_LIMIT", ValueFrom: &corev1.EnvVarSource{
					ResourceFieldRef: &corev1.ResourceFieldSelector{
						Divisor:  diskDivisor,
						Resource: "limits.ephemeral-storage",
					},
				}},
				{Name: "LANG", Value: "en_US.UTF-8"},
				// json.Marshal writes values in sorted key order
				{Name: "VCAP_APPLICATION", Value: `{"application_id":"12345","application_name":"my-app","application_uris":["my-app.example.com"],"limits":{"disk":$(DISK_LIMIT),"mem":$(MEMORY_LIMIT)},"name":"my-app","process_id":"12345","process_type":"web","space_name":"my-ns","uris":["my-app.example.com"]}`},
				{Name: "VCAP_SERVICES", ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "VCAP_SERVICES",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "kf-injected-envs-my-app",
						},
						Optional: ptr.Bool(false),
					},
				}},
			},
		},
		"task app": {
			app:     app,
			runtime: CFTask,
			wantOut: []corev1.EnvVar{
				{Name: "CF_INSTANCE_IP", ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				}},
				{Name: "CF_INSTANCE_INTERNAL_IP", Value: "$(CF_INSTANCE_IP)"},
				{Name: "VCAP_APP_HOST", Value: "$(CF_INSTANCE_IP)"},
				{Name: "CF_INSTANCE_PORT", Value: "9999"},
				{Name: "CF_INSTANCE_ADDR", Value: "$(CF_INSTANCE_IP):$(CF_INSTANCE_PORT)"},
				{Name: "CF_INSTANCE_GUID", ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.uid",
					},
				}},
				{Name: "INSTANCE_GUID", Value: "$(CF_INSTANCE_GUID)"},
				{Name: "MEMORY_LIMIT", ValueFrom: &corev1.EnvVarSource{
					ResourceFieldRef: &corev1.ResourceFieldSelector{
						Divisor:  memoryDivisor,
						Resource: "limits.memory",
					},
				}},
				{Name: "DISK_LIMIT", ValueFrom: &corev1.EnvVarSource{
					ResourceFieldRef: &corev1.ResourceFieldSelector{
						Divisor:  diskDivisor,
						Resource: "limits.ephemeral-storage",
					},
				}},
				{Name: "LANG", Value: "en_US.UTF-8"},
				// json.Marshal writes values in sorted key order
				{Name: "VCAP_APPLICATION", Value: `{"application_id":"12345","application_name":"my-app","application_uris":["my-app.example.com"],"limits":{"disk":$(DISK_LIMIT),"mem":$(MEMORY_LIMIT)},"name":"my-app","process_id":"12345","process_type":"web","space_name":"my-ns","uris":["my-app.example.com"]}`},
				{Name: "VCAP_SERVICES", ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "VCAP_SERVICES",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "kf-injected-envs-my-app",
						},
						Optional: ptr.Bool(false),
					},
				}},
				{Name: "DATABASE_URL", ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "DATABASE_URL",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "kf-injected-envs-my-app",
						},
						Optional: ptr.Bool(true),
					},
				}},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gotOut := BuildRuntimeEnvVars(tc.runtime, tc.app)

			testutil.AssertEqual(t, "environment", tc.wantOut, gotOut)
		})
	}
}

func TestRuntimeEnvVarList(t *testing.T) {
	cases := map[string]struct {
		runtime EnvRuntime
	}{
		"running": {runtime: CFRunning},
		"staging": {runtime: CFStaging},
		"task":    {runtime: CFTask},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			expectedVars := sets.NewString()
			for _, v := range BuildRuntimeEnvVars(tc.runtime, &v1alpha1.App{}) {
				expectedVars.Insert(v.Name)
			}

			actualList := RuntimeEnvVarList(tc.runtime)

			testutil.AssertEqual(t, "variable list matches actual", expectedVars, actualList)
		})
	}
}

func TestRuntimeEnvVarDocs(t *testing.T) {
	cases := map[string]struct {
		runtime EnvRuntime
	}{
		"running": {runtime: CFRunning},
		"staging": {runtime: CFStaging},
		"task":    {runtime: CFTask},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			// sanity check that all the expected vars are there
			expectedVars := RuntimeEnvVarList(tc.runtime).List()
			docs := RuntimeEnvVarDocs(tc.runtime)
			testutil.AssertContainsAll(t, docs, expectedVars)
		})
	}
}
