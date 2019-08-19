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
	"fmt"
	"testing"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ExampleBuildpackBuildImageDestination() {
	app := &v1alpha1.App{}
	app.Name = "myapp"
	app.Namespace = "myspace"
	app.Spec.Source.UpdateRequests = 0xfacade

	space := &v1alpha1.Space{}
	space.Spec.BuildpackBuild.ContainerRegistry = "gcr.io/my-project"

	fmt.Println(BuildpackBuildImageDestination(app, space))

	// Output: gcr.io/my-project/app_myspace_myapp:facade
}

func ExampleBuildpackBuildImageDestination_noRegistry() {
	app := &v1alpha1.App{}
	app.Name = "myapp"
	app.Namespace = "myspace"
	app.Spec.Source.UpdateRequests = 0xfacade

	fmt.Println(BuildpackBuildImageDestination(app, &v1alpha1.Space{}))

	// Output: app_myspace_myapp:facade
}

func TestMakeSource(t *testing.T) {

	space := v1alpha1.Space{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myspace",
		},
		Spec: v1alpha1.SpaceSpec{
			Security: v1alpha1.SpaceSpecSecurity{
				BuildServiceAccount: "build-service-account",
			},

			BuildpackBuild: v1alpha1.SpaceSpecBuildpackBuild{
				ContainerRegistry: "gcr.io/dest",
			},
		},
	}

	appObjectMeta := metav1.ObjectMeta{
		Name:      "mybuildpackapp",
		Namespace: "myspace",
	}

	appOwnerRef := []metav1.OwnerReference{
		{
			APIVersion:         "kf.dev/v1alpha1",
			Kind:               "App",
			Name:               "mybuildpackapp",
			Controller:         boolPtr(true),
			BlockOwnerDeletion: boolPtr(true),
		},
	}

	cases := map[string]struct {
		app   v1alpha1.App
		space v1alpha1.Space

		expected v1alpha1.Source
	}{
		"buildpack": {
			app: v1alpha1.App{
				ObjectMeta: appObjectMeta,
				Spec: v1alpha1.AppSpec{
					Source: v1alpha1.SourceSpec{
						UpdateRequests: 0xdeadbeef,
						BuildpackBuild: v1alpha1.SourceSpecBuildpackBuild{
							Source: "gcr.io/my-source-image:latest",
						},
					},
				},
			},
			space: space,

			expected: v1alpha1.Source{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mybuildpackapp-deadbeef",
					Namespace: "myspace",
					Labels: map[string]string{
						"app.kubernetes.io/component":  "build",
						"app.kubernetes.io/managed-by": "kf",
						"app.kubernetes.io/name":       "mybuildpackapp",
					},
					OwnerReferences: appOwnerRef,
				},
				Spec: v1alpha1.SourceSpec{
					UpdateRequests: 0xdeadbeef,
					ServiceAccount: "build-service-account",
					BuildpackBuild: v1alpha1.SourceSpecBuildpackBuild{
						Source: "gcr.io/my-source-image:latest",
						Image:  "gcr.io/dest/app_myspace_mybuildpackapp:deadbeef",
					},
				},
			},
		},
		"docker": {
			app: v1alpha1.App{
				ObjectMeta: appObjectMeta,
				Spec: v1alpha1.AppSpec{
					Source: v1alpha1.SourceSpec{
						UpdateRequests: 0xfacade,
						ContainerImage: v1alpha1.SourceSpecContainerImage{
							Image: "mysql/mysql:v1",
						},
					},
				},
			},
			space: space,

			expected: v1alpha1.Source{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mybuildpackapp-facade",
					Namespace: "myspace",
					Labels: map[string]string{
						"app.kubernetes.io/component":  "build",
						"app.kubernetes.io/managed-by": "kf",
						"app.kubernetes.io/name":       "mybuildpackapp",
					},
					OwnerReferences: appOwnerRef,
				},
				Spec: v1alpha1.SourceSpec{
					UpdateRequests: 0xfacade,
					ServiceAccount: "build-service-account",
					ContainerImage: v1alpha1.SourceSpecContainerImage{
						Image: "mysql/mysql:v1",
					},
				},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual, err := MakeSource(&tc.app, &tc.space)
			testutil.AssertNil(t, "MakeSource error", err)

			testutil.AssertEqual(t, "Source", &tc.expected, actual)
		})
	}
}

func boolPtr(b bool) *bool {
	tmp := &b
	return tmp
}
