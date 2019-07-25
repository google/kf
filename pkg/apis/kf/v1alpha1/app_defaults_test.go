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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestAppSpec_SetDefaults_BlankContainer(t *testing.T) {
	t.Parallel()

	app := &App{}
	app.SetDefaults(context.Background())

	testutil.AssertEqual(t, "len(spec.template.spec.containers)", 1, len(app.Spec.Template.Spec.Containers))
	testutil.AssertEqual(t, "spec.template.spec.containers.name", "", app.Spec.Template.Spec.Containers[0].Name)
}

func TestAppSpec_SetDefaults_ResourceLimits_Default(t *testing.T) {
	t.Parallel()

	app := &App{
		Spec: AppSpec{
			Template: AppSpecTemplate{
				Spec: v1.PodSpec{
					Containers: []v1.Container{{}},
				},
			},
		},
	}
	app.SetDefaults(context.Background())

	wantMem := resource.MustParse("1Gi")
	wantStorage := resource.MustParse("1Gi")
	wantCPU := resource.MustParse("1")

	testutil.AssertEqual(t, "default memory request", wantMem, app.Spec.Template.Spec.Containers[0].Resources.Requests[v1.ResourceMemory])
	testutil.AssertEqual(t, "default storage request", wantStorage, app.Spec.Template.Spec.Containers[0].Resources.Requests[v1.ResourceEphemeralStorage])
	testutil.AssertEqual(t, "default CPU request", wantCPU, app.Spec.Template.Spec.Containers[0].Resources.Requests[v1.ResourceCPU])
}

func TestAppSpec_SetDefaults_ResourceLimits_AlreadySet(t *testing.T) {
	t.Parallel()

	wantMem := resource.MustParse("2Gi")
	wantStorage := resource.MustParse("2Gi")
	wantCPU := resource.MustParse("2")

	app := &App{
		Spec: AppSpec{
			Template: AppSpecTemplate{
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceMemory:           wantMem,
								v1.ResourceEphemeralStorage: wantStorage,
								v1.ResourceCPU:              wantCPU,
							},
						},
					}},
				},
			},
		},
	}

	app.SetDefaults(context.Background())

	testutil.AssertEqual(t, "default memory request", wantMem, app.Spec.Template.Spec.Containers[0].Resources.Requests[v1.ResourceMemory])
	testutil.AssertEqual(t, "default storage request", wantStorage, app.Spec.Template.Spec.Containers[0].Resources.Requests[v1.ResourceEphemeralStorage])
	testutil.AssertEqual(t, "default CPU request", wantCPU, app.Spec.Template.Spec.Containers[0].Resources.Requests[v1.ResourceCPU])
}
