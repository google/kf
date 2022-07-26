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
	"context"
	"fmt"
	"testing"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ExampleKfApp() {
	space := NewKfApp()
	// Setup
	space.SetName("nsname")

	// Values
	fmt.Println(space.GetName())

	// Output: nsname
}

func TestKfApp_ToApp(t *testing.T) {
	app := NewKfApp()
	app.SetName("foo")
	actual := app.ToApp()

	expected := &v1alpha1.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "kf.dev/v1alpha1",
		},
		Spec: v1alpha1.AppSpec{
			Template: v1alpha1.AppSpecTemplate{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
		},
	}
	expected.Name = "foo"

	testutil.AssertEqual(t, "generated service", expected, actual)
}

func ExampleKfApp_GetEnvVars() {
	myApp := NewKfApp()
	myApp.SetEnvVars([]corev1.EnvVar{
		{Name: "FOO", Value: "2"},
		{Name: "BAR", Value: "0"},
	})

	env := myApp.GetEnvVars()

	for _, e := range env {
		fmt.Println("Key", e.Name, "Value", e.Value)
	}

	// Output: Key FOO Value 2
	// Key BAR Value 0
}

func ExampleKfApp_GetEnvVars_emptyApp() {
	myApp := NewKfApp()

	env := myApp.GetEnvVars()

	fmt.Println(env)

	// Output: []
}

func ExampleKfApp_MergeEnvVars() {
	myApp := NewKfApp()
	myApp.SetEnvVars([]corev1.EnvVar{
		{Name: "FOO", Value: "0"},
		{Name: "BAR", Value: "0"},
	})

	myApp.MergeEnvVars([]corev1.EnvVar{
		{Name: "FOO", Value: "1"},  // will replace old
		{Name: "BAZZ", Value: "0"}, // will be added
	})

	env := myApp.GetEnvVars()

	for _, e := range env {
		fmt.Println("Key", e.Name, "Value", e.Value)
	}

	// Output: Key BAR Value 0
	// Key BAZZ Value 0
	// Key FOO Value 1
}

func ExampleKfApp_DeleteEnvVars() {
	myApp := NewKfApp()
	myApp.SetEnvVars([]corev1.EnvVar{
		{Name: "FOO", Value: "0"},
		{Name: "BAR", Value: "0"},
	})

	myApp.DeleteEnvVars([]string{"FOO", "DOES_NOT_EXIST"})

	for _, e := range myApp.GetEnvVars() {
		fmt.Println("Key", e.Name, "Value", e.Value)
	}

	// Output: Key BAR Value 0
}

func ExampleKfApp_GetNamespace() {
	myApp := NewKfApp()
	myApp.SetNamespace("my-ns")

	fmt.Println(myApp.GetNamespace())

	// Output: my-ns
}

func rwb(host, domain, path string) v1alpha1.RouteWeightBinding {
	return v1alpha1.RouteWeightBinding{
		RouteSpecFields: v1alpha1.RouteSpecFields{
			Hostname: host,
			Domain:   domain,
			Path:     path,
		},
	}
}

func TestKfApp_MergeRoute(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		routes     []v1alpha1.RouteWeightBinding
		merge      v1alpha1.RouteWeightBinding
		wantRoutes []v1alpha1.RouteWeightBinding
	}{
		"empty start": {
			routes: nil,
			merge:  rwb("test", "example.com", "/"),
			wantRoutes: []v1alpha1.RouteWeightBinding{
				rwb("test", "example.com", "/"),
			},
		},
		"appends new": {
			routes: []v1alpha1.RouteWeightBinding{
				rwb("test", "example.com", "/foo"),
			},
			merge: rwb("test", "example.com", "/bar"),
			wantRoutes: []v1alpha1.RouteWeightBinding{
				rwb("test", "example.com", "/foo"),
				rwb("test", "example.com", "/bar"),
			},
		},
		"removes duplicates": {
			routes: []v1alpha1.RouteWeightBinding{
				rwb("test", "example.com", "/foo"),
				rwb("test", "example.com", "/foo"),
			},
			merge: rwb("test", "example.com", "/foo"),
			wantRoutes: []v1alpha1.RouteWeightBinding{
				rwb("test", "example.com", "/foo"),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			app := NewKfApp()
			app.Spec.Routes = tc.routes

			app.MergeRoute(tc.merge)

			testutil.AssertEqual(t, "after", tc.wantRoutes, app.Spec.Routes)
		})
	}
}

func TestKfApp_RemoveRoute(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		routes     []v1alpha1.RouteWeightBinding
		remove     v1alpha1.RouteWeightBinding
		wantRoutes []v1alpha1.RouteWeightBinding
	}{
		"from nil routes": {
			routes:     nil,
			remove:     rwb("test", "example.com", "/"),
			wantRoutes: nil,
		},
		"does not exist": {
			routes: []v1alpha1.RouteWeightBinding{
				rwb("a", "example.com", "/"),
				rwb("b", "example.com", "/"),
			},
			remove: rwb("c", "example.com", "/"),
			wantRoutes: []v1alpha1.RouteWeightBinding{
				rwb("a", "example.com", "/"),
				rwb("b", "example.com", "/"),
			},
		},
		"multiple": {
			routes: []v1alpha1.RouteWeightBinding{
				rwb("c", "example.com", "/"),
				rwb("b", "example.com", "/"),
				rwb("c", "example.com", "/"),
			},
			remove: rwb("c", "example.com", "/"),
			wantRoutes: []v1alpha1.RouteWeightBinding{
				rwb("b", "example.com", "/"),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			app := NewKfApp()
			app.Spec.Routes = tc.routes

			app.RemoveRoute(context.Background(), tc.remove)

			testutil.AssertEqual(t, "after", tc.wantRoutes, app.Spec.Routes)
		})
	}
}

func TestKfApp_RemoveRoutesForClaim(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		routes     []v1alpha1.RouteWeightBinding
		claim      v1alpha1.RouteSpecFields
		wantRoutes []v1alpha1.RouteWeightBinding
	}{
		"from nil routes": {
			routes:     nil,
			claim:      rwb("test", "example.com", "/").RouteSpecFields,
			wantRoutes: nil,
		},
		"does not exist": {
			routes: []v1alpha1.RouteWeightBinding{
				rwb("a", "example.com", "/"),
				rwb("b", "example.com", "/"),
			},
			claim: rwb("c", "example.com", "/").RouteSpecFields,
			wantRoutes: []v1alpha1.RouteWeightBinding{
				rwb("a", "example.com", "/"),
				rwb("b", "example.com", "/"),
			},
		},
		"multiple": {
			routes: []v1alpha1.RouteWeightBinding{
				rwb("c", "example.com", "/"),
				rwb("b", "example.com", "/"),
				rwb("c", "example.com", "/"),
			},
			claim: rwb("c", "example.com", "/").RouteSpecFields,
			wantRoutes: []v1alpha1.RouteWeightBinding{
				rwb("b", "example.com", "/"),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			app := NewKfApp()
			app.Spec.Routes = tc.routes

			app.RemoveRoutesForClaim(tc.claim)

			testutil.AssertEqual(t, "after", tc.wantRoutes, app.Spec.Routes)
		})
	}
}

func TestKfApp_HasMatchingRoutes(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		routes []v1alpha1.RouteWeightBinding
		claim  v1alpha1.RouteSpecFields
		want   bool
	}{
		"from nil routes": {
			routes: nil,
			claim:  rwb("test", "example.com", "/").RouteSpecFields,
			want:   false,
		},
		"does not exist": {
			routes: []v1alpha1.RouteWeightBinding{
				rwb("a", "example.com", "/"),
				rwb("b", "example.com", "/"),
			},
			claim: rwb("c", "example.com", "/").RouteSpecFields,
			want:  false,
		},
		"multiple": {
			routes: []v1alpha1.RouteWeightBinding{
				rwb("c", "example.com", "/"),
				rwb("b", "example.com", "/"),
				rwb("c", "example.com", "/"),
			},
			claim: rwb("c", "example.com", "/").RouteSpecFields,
			want:  true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			app := NewKfApp()
			app.Spec.Routes = tc.routes

			got := app.HasMatchingRoutes(tc.claim)

			testutil.AssertEqual(t, "matched", tc.want, got)
		})
	}
}
