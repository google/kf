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
	"fmt"
	"testing"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
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

func ExampleKfApp_GetHealthCheck() {
	app := NewKfApp()
	app.SetContainer(corev1.Container{
		ReadinessProbe: &corev1.Probe{
			TimeoutSeconds: 34,
		},
	})

	fmt.Println("Timeout:", app.GetHealthCheck().TimeoutSeconds)

	// Output: Timeout: 34
}

func ExampleKfApp_GetClusterURL() {
	app := NewKfApp()
	app.Status.Address = &duckv1alpha1.Addressable{
		Addressable: duckv1beta1.Addressable{
			URL: &apis.URL{
				Host:   "app-a.some-namespace.svc.cluster.local",
				Scheme: "http",
			},
		},
	}

	fmt.Println(app.GetClusterURL())

	// Output: http://app-a.some-namespace.svc.cluster.local
}
