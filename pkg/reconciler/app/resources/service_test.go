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

	kfv1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ExamplePodLabels() {
	labels := PodLabels(&kfv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-app",
			Namespace: "my-ns",
		},
	})
	fmt.Println(labels)

	// Output: map[app.kubernetes.io/component:app-server app.kubernetes.io/managed-by:kf app.kubernetes.io/name:my-app]
}

func ExampleServiceName() {
	app := &kfv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-app",
			Namespace: "my-ns",
		},
	}

	fmt.Println(ServiceName(app))

	// Output: my-app
}

func ExampleServiceNameForAppName() {
	fmt.Println(ServiceNameForAppName("my-app"))

	// Output: my-app
}

func TestMakeService(t *testing.T) {
	appWithoutPorts := &kfv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "my-ns",
		},
		Spec: kfv1alpha1.AppSpec{
			Template: kfv1alpha1.AppSpecTemplate{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "user-service",
							// no particular port info
						},
					},
				},
			},
		},
	}

	cases := map[string]struct {
		app *kfv1alpha1.App
	}{
		"default": {
			app: appWithoutPorts.DeepCopy(),
		},
		"custom-port": {
			app: (func() *kfv1alpha1.App {
				tmp := appWithoutPorts.DeepCopy()
				tmp.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
					{Name: "http-9999", ContainerPort: 9999},
				}
				return tmp
			})(),
		},
		"port 80 already defined": {
			app: (func() *kfv1alpha1.App {
				tmp := appWithoutPorts.DeepCopy()
				tmp.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
					{Name: "http-80", ContainerPort: 80},
				}
				return tmp
			})(),
		},
		"UserPortName already defined": {
			app: (func() *kfv1alpha1.App {
				tmp := appWithoutPorts.DeepCopy()
				tmp.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
					{Name: UserPortName, ContainerPort: 8888},
				}
				return tmp
			})(),
		},
		"multiple custom ports": {
			app: (func() *kfv1alpha1.App {
				tmp := appWithoutPorts.DeepCopy()
				tmp.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
					{Name: "http-8888", ContainerPort: 8888},
					{Name: "tcp-90", ContainerPort: 90},
					{Name: "http-8000", ContainerPort: 8000},
				}
				return tmp
			})(),
		},
		"custom ports mixed names": {
			app: (func() *kfv1alpha1.App {
				tmp := appWithoutPorts.DeepCopy()
				tmp.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
					{ContainerPort: 8888},
					{ContainerPort: 90},
					{Name: "ssh", ContainerPort: 22},
				}
				return tmp
			})(),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			svc := MakeService(tc.app)

			testutil.AssertGoldenJSONContext(t, "service", svc, map[string]interface{}{
				"app": tc.app,
			})
		})
	}
}

func Test_getUserPort(t *testing.T) {
	type args struct {
		app *kfv1alpha1.App
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUserPort(tt.args.app); got != tt.want {
				t.Errorf("getUserPort() = %v, want %v", got, tt.want)
			}
		})
	}
}
