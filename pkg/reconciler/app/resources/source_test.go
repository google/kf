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

func TestMakeSourceLabels(t *testing.T) {
	app := &v1alpha1.App{}
	app.Name = "my-app"

	// hard-code expected here to ensure it doesn't change, if it changes
	// we could suddenly spawn up hundreds of buils all at once when a user
	// upgrades
	expected := map[string]string{
		"app.kubernetes.io/component":  "build",
		"app.kubernetes.io/managed-by": "kf",
		"app.kubernetes.io/name":       "my-app",
	}

	actual := MakeSourceLabels(app)
	testutil.AssertEqual(t, "labels", expected, actual)
}

func ExampleBuildpackBulidImageDestination() {
	app := &v1alpha1.App{}
	app.Name = "myapp"
	app.Namespace = "myspace"

	space := &v1alpha1.Space{}
	space.Spec.BuildpackBuild.ContainerRegistry = "gcr.io/my-project"

	fmt.Println(BuildpackBulidImageDestination(app, space, 22453731916))

	// Output: gcr.io/my-project/app-myspace-myapp:abcdefg
}

func ExampleBuildpackBulidImageDestination_noRegistry() {
	app := &v1alpha1.App{}
	app.Name = "myapp"
	app.Namespace = "myspace"

	fmt.Println(BuildpackBulidImageDestination(app, &v1alpha1.Space{}, 22453731916))

	// Output: app-myspace-myapp:abcdefg
}

func TestMakeSource(t *testing.T) {
	abcdefg := int64(22453731916)

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
		app    v1alpha1.App
		space  v1alpha1.Space
		suffix int64

		expected v1alpha1.Source
	}{
		"buildpack": {
			app: v1alpha1.App{
				ObjectMeta: appObjectMeta,
				Spec: v1alpha1.AppSpec{
					Source: v1alpha1.SourceSpec{
						BuildpackBuild: v1alpha1.SourceSpecBuildpackBuild{
							Source: "gcr.io/my-source-image:latest",
						},
					},
				},
			},
			space:  space,
			suffix: abcdefg,

			expected: v1alpha1.Source{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mybuildpackapp-abcdefg",
					Namespace: "myspace",
					Labels: map[string]string{
						"app.kubernetes.io/component":  "build",
						"app.kubernetes.io/managed-by": "kf",
						"app.kubernetes.io/name":       "mybuildpackapp",
					},
					OwnerReferences: appOwnerRef,
				},
				Spec: v1alpha1.SourceSpec{
					ServiceAccount: "build-service-account",
					BuildpackBuild: v1alpha1.SourceSpecBuildpackBuild{
						Source: "gcr.io/my-source-image:latest",
						Image:  "gcr.io/dest/app-myspace-mybuildpackapp:abcdefg",
					},
				},
			},
		},
		"docker": {
			app: v1alpha1.App{
				ObjectMeta: appObjectMeta,
				Spec: v1alpha1.AppSpec{
					Source: v1alpha1.SourceSpec{
						ContainerImage: v1alpha1.SourceSpecContainerImage{
							Image: "mysql/mysql:v1",
						},
					},
				},
			},
			space:  space,
			suffix: abcdefg,

			expected: v1alpha1.Source{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mybuildpackapp-abcdefg",
					Namespace: "myspace",
					Labels: map[string]string{
						"app.kubernetes.io/component":  "build",
						"app.kubernetes.io/managed-by": "kf",
						"app.kubernetes.io/name":       "mybuildpackapp",
					},
					OwnerReferences: appOwnerRef,
				},
				Spec: v1alpha1.SourceSpec{
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
			actual, err := MakeSource(&tc.app, &tc.space, tc.suffix)
			testutil.AssertNil(t, "MakeSource error", err)

			testutil.AssertEqual(t, "Source", &tc.expected, actual)
		})
	}
}

func boolPtr(b bool) *bool {
	tmp := &b
	return tmp
}
