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

package resources

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/google/kf/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func ExampleMakeServiceBindingLabels() {
	app := &v1alpha1.App{}
	app.Name = "my-app"
	binding := &v1alpha1.AppSpecServiceBinding{
		Instance:    "my-service",
		BindingName: "cool-binding",
	}
	app.Spec.ServiceBindings = []v1alpha1.AppSpecServiceBinding{*binding}

	labels := MakeServiceBindingLabels(app, binding)
	describe.Labels(os.Stdout, labels)

	// Output: app.kubernetes.io/component=servicebinding
	// app.kubernetes.io/managed-by=kf
	// app.kubernetes.io/name=my-app
	// bindings.kf.dev/app-name=my-app
	// bindings.kf.dev/binding-name=cool-binding
}

func ExampleMakeServiceBindingName() {
	app := &v1alpha1.App{}
	app.Name = "my-app"
	binding := &v1alpha1.AppSpecServiceBinding{
		Instance:    "my-service",
		BindingName: "a-cool-binding",
	}
	app.Spec.ServiceBindings = []v1alpha1.AppSpecServiceBinding{*binding}

	fmt.Println(MakeServiceBindingName(app, binding))

	// Output: kf-binding-my-app-a-cool-binding
}

func TestMakeServiceBindingAppSelector(t *testing.T) {
	t.Parallel()

	s := MakeServiceBindingAppSelector("my-app")

	good := labels.Set{
		"bindings.kf.dev/app-name": "my-app",
	}
	bad := labels.Set{
		"bindings.kf.dev/app-name": "not-my-app",
	}

	testutil.AssertEqual(t, "matches", true, s.Matches(good))
	testutil.AssertEqual(t, "doesn't match", false, s.Matches(bad))
}

func TestMakeServiceBinding(t *testing.T) {
	appSpecBinding := &v1alpha1.AppSpecServiceBinding{
		Instance:    "my-service-instance",
		BindingName: "a-cool-binding",
		Parameters:  []byte(`{"username":"me"}`),
	}
	app := &v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-app",
			Namespace: "my-namespace",
		},
		Spec: v1alpha1.AppSpec{
			ServiceBindings: []v1alpha1.AppSpecServiceBinding{*appSpecBinding},
		},
	}

	binding, err := MakeServiceBinding(app, appSpecBinding)
	testutil.AssertNil(t, "error", err)
	testutil.AssertEqual(t, "name", "kf-binding-my-app-a-cool-binding", binding.Name)
	testutil.AssertEqual(t, "namespace", "my-namespace", binding.Namespace)
	testutil.AssertEqual(t, "instance name", "my-service-instance", binding.Spec.InstanceRef.Name)
	testutil.AssertEqual(t, "parameters", `{"username":"me"}`, string(binding.Spec.Parameters.Raw))

	expectedLabels := map[string]string{
		"app.kubernetes.io/component":  "servicebinding",
		"app.kubernetes.io/managed-by": "kf",
		"app.kubernetes.io/name":       "my-app",
		"bindings.kf.dev/app-name":     "my-app",
		"bindings.kf.dev/binding-name": "a-cool-binding",
	}

	testutil.AssertEqual(t, "labels", expectedLabels, binding.Labels)
}
