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
	"testing"
	"time"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/reconciler/build/config"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ExampleDestinationImageName() {
	build := &v1alpha1.Build{}
	build.UID = "9463ad9e-4dd8-11ea-91a4-42010a80008d"
	build.Name = "myapp"
	build.Namespace = "myspace"

	space := &v1alpha1.Space{}
	space.Status.BuildConfig.ContainerRegistry = "gcr.io/my-project"

	fmt.Println(DestinationImageName(build, space))

	// Output: gcr.io/my-project/app_myspace_myapp:9463ad9e-4dd8-11ea-91a4-42010a80008d
}

func ExampleDestinationImageName_noRegistry() {
	build := &v1alpha1.Build{}
	build.UID = "32150eb9-8941-4ac9-ba23-522d902e8b81"
	build.Name = "myapp"
	build.Namespace = "myspace"

	fmt.Println(DestinationImageName(build, &v1alpha1.Space{}))

	// Output: app_myspace_myapp:32150eb9-8941-4ac9-ba23-522d902e8b81
}

func ExampleTaskRunName() {
	build := &v1alpha1.Build{}
	build.Name = "my-build"

	fmt.Println(TaskRunName(build))

	// Output: my-build
}

func exampleCustomTaskBuild() (*v1alpha1.Build, *tektonv1beta1.TaskSpec) {
	build := &v1alpha1.Build{}
	build.Name = "my-build"
	build.UID = "0d5c53ff-edf1-4d42-8d1a-fdd5b5cf23d3"
	build.Namespace = "my-namespace"
	build.Spec.Name = "buildpack"
	build.Spec.Kind = "ClusterTask"

	build.Spec.Params = []v1alpha1.BuildParam{
		{
			Name:  "some-param",
			Value: "some-value",
		},
	}

	build.Spec.Env = []corev1.EnvVar{
		{Name: "KEY", Value: "value"},
	}

	build.Spec.NodeSelector = map[string]string{
		"os": "debian4.9",
	}

	taskSpec := &tektonv1beta1.TaskSpec{}

	return build, taskSpec
}

func ExampleMakeTaskRun_customTask_taskRun_addedParams() {
	build, taskSpec := exampleCustomTaskBuild()

	taskSpec.Params = append(
		taskSpec.Params,
		tektonv1beta1.ParamSpec{
			Name:        v1alpha1.TaskRunParamSourceImage,
			Type:        tektonv1beta1.ParamTypeString,
			Description: "",
			Default:     nil,
		},
		tektonv1beta1.ParamSpec{
			Name:        v1alpha1.TaskRunParamDestinationImage,
			Type:        tektonv1beta1.ParamTypeString,
			Description: "",
			Default:     tektonv1beta1.NewArrayOrString("gcr.io/test"),
		},
	)

	nodeSelectorMap := map[string]string{
		"disktype": "ssd",
		"cpu":      "amd64",
	}

	space := &v1alpha1.Space{
		Status: v1alpha1.SpaceStatus{
			BuildConfig: v1alpha1.SpaceStatusBuildConfig{
				ServiceAccount: "some-account",
			},
		},
		Spec: v1alpha1.SpaceSpec{
			RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{
				NodeSelector: nodeSelectorMap,
			},
		},
	}

	taskRun, err := MakeTaskRun(build, taskSpec, space, nil, nil, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Name:", taskRun.Name)
	fmt.Println("Label Count:", len(taskRun.Labels))
	fmt.Println("Managed By:", taskRun.Labels[managedByLabel])
	fmt.Println("NetworkPolicy:", taskRun.Labels[v1alpha1.NetworkPolicyLabel])
	fmt.Println("Service Account:", taskRun.Spec.ServiceAccountName)
	fmt.Println("OwnerReferences Count:", len(taskRun.OwnerReferences))
	fmt.Println("Input Count:", len(taskRun.Spec.Params))
	fmt.Println("Env Count:", len(taskRun.Spec.TaskSpec.StepTemplate.Env))
	fmt.Println("Node Selector Count:", len(taskRun.Spec.PodTemplate.NodeSelector))

	// Output: Name: my-build
	// Label Count: 2
	// Managed By: kf
	// NetworkPolicy: build
	// Service Account: some-account
	// OwnerReferences Count: 1
	// Input Count: 2
	// Env Count: 1
	// Node Selector Count: 3
}

func TestMakeTaskRun(t *testing.T) {
	t.Parallel()

	findParam := func(t *testing.T, tr *tektonv1beta1.TaskRun, param string) (string, bool) {
		var value string
		var found bool
		for _, p := range tr.Spec.Params {
			if p.Name == param {
				// Instead of just returning here, keep going. Fail if
				// multiple are found.
				if found {
					t.Fatalf("multiple %q params found", param)
				}
				found = true
				value = p.Value.StringVal
			}
		}

		return value, found
	}

	t.Run("keep existing source image when not overridden", func(t *testing.T) {
		tr, err := MakeTaskRun(&v1alpha1.Build{
			Spec: v1alpha1.BuildSpec{
				Params: []v1alpha1.BuildParam{
					{Name: v1alpha1.SourceImageParamName, Value: "some-image"},
				},
			},
		}, &tektonv1beta1.TaskSpec{}, &v1alpha1.Space{}, nil, nil, nil)

		testutil.AssertNil(t, "err", err)
		got, _ := findParam(t, tr, v1alpha1.SourceImageParamName)
		testutil.AssertEqual(t, "image", "some-image", got)
		_, ok := findParam(t, tr, v1alpha1.SourcePackageNameParamName)
		testutil.AssertFalse(t, v1alpha1.SourcePackageNameParamName, ok)
	})

	t.Run("set source package params when specified", func(t *testing.T) {
		tr, err := MakeTaskRun(
			&v1alpha1.Build{},
			&tektonv1beta1.TaskSpec{},
			&v1alpha1.Space{},
			&v1alpha1.SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-name",
					Namespace: "some-namespace",
				},
			},
			nil,
			nil,
		)

		testutil.AssertNil(t, "err", err)
		_, ok := findParam(t, tr, v1alpha1.SourceImageParamName)
		testutil.AssertFalse(t, v1alpha1.SourceImageParamName, ok)
		got, _ := findParam(t, tr, v1alpha1.SourcePackageNameParamName)
		testutil.AssertEqual(t, v1alpha1.SourcePackageNameParamName, "some-name", got)
		got, _ = findParam(t, tr, v1alpha1.SourcePackageNamespaceParamName)
		testutil.AssertEqual(t, v1alpha1.SourcePackageNamespaceParamName, "some-namespace", got)
	})

	t.Run("Init step when WI used and it is a V3 pipeline", func(t *testing.T) {
		tr, err := MakeTaskRun(
			&v1alpha1.Build{
				Spec: v1alpha1.BuildSpec{
					BuildTaskRef: v1alpha1.BuildTaskRef{
						Name: v1alpha1.BuildpackV3BuildTaskName,
					},
				},
			},
			&tektonv1beta1.TaskSpec{},
			&v1alpha1.Space{},
			&v1alpha1.SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-name",
					Namespace: "some-namespace",
				},
			},
			&config.SecretsConfig{
				GoogleServiceAccount: "GSA",
			},
			nil,
		)

		testutil.AssertNil(t, "err", err)
		testutil.AssertEqual(t, "init step name", "wait-for-wi", tr.Spec.TaskSpec.Steps[0].Name)
	})

	t.Run("Set the istio injection annotation to false", func(t *testing.T) {
		tr, err := MakeTaskRun(
			&v1alpha1.Build{},
			&tektonv1beta1.TaskSpec{},
			&v1alpha1.Space{},
			&v1alpha1.SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-name",
					Namespace: "some-namespace",
				},
			},
			nil,
			&kfconfig.DefaultsConfig{
				BuildDisableIstioSidecar: true,
			},
		)

		testutil.AssertNil(t, "err", err)
		testutil.AssertEqual(t, "annotation", "false", tr.Annotations["sidecar.istio.io/inject"])
	})

	t.Run("Set the istio injection annotation to true", func(t *testing.T) {
		tr, err := MakeTaskRun(
			&v1alpha1.Build{},
			&tektonv1beta1.TaskSpec{},
			&v1alpha1.Space{},
			&v1alpha1.SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-name",
					Namespace: "some-namespace",
				},
			},
			nil,
			&kfconfig.DefaultsConfig{
				BuildDisableIstioSidecar: false,
			},
		)

		testutil.AssertNil(t, "err", err)
		testutil.AssertEqual(t, "annotation", "true", tr.Annotations["sidecar.istio.io/inject"])
	})

	t.Run("Set default timeout when configured timeout is empty", func(t *testing.T) {
		tr, err := MakeTaskRun(&v1alpha1.Build{},
			&tektonv1beta1.TaskSpec{},
			&v1alpha1.Space{},
			&v1alpha1.SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-name",
					Namespace: "some-namespace",
				},
			},
			nil, nil)
		testutil.AssertNil(t, "err", err)
		testutil.AssertEqual(t, "timeout", time.Duration(1*time.Hour).Seconds(), tr.Spec.Timeout.Duration.Seconds())
	})

	t.Run("Set timeout with configured timeout", func(t *testing.T) {
		tr, err := MakeTaskRun(&v1alpha1.Build{},
			&tektonv1beta1.TaskSpec{},
			&v1alpha1.Space{},
			&v1alpha1.SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-name",
					Namespace: "some-namespace",
				},
			},
			nil, &kfconfig.DefaultsConfig{
				BuildTimeout: "30m",
			},
		)
		testutil.AssertNil(t, "err", err)
		testutil.AssertEqual(t, "timeout", time.Duration(30*time.Minute).Seconds(), tr.Spec.Timeout.Duration.Seconds())
	})

	t.Run("Set build node selectors with configured buildNodeSelectors", func(t *testing.T) {
		testBuildNodeSelectors := make(map[string]string)
		testBuildNodeSelectors["disktype"] = "ssd"

		tr, err := MakeTaskRun(&v1alpha1.Build{},
			&tektonv1beta1.TaskSpec{},
			&v1alpha1.Space{},
			&v1alpha1.SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-name",
					Namespace: "some-namespace",
				},
			},
			nil, &kfconfig.DefaultsConfig{
				BuildNodeSelectors: testBuildNodeSelectors,
			},
		)
		testutil.AssertNil(t, "err", err)
		testutil.AssertEqual(t, "buildNodeSelectors", testBuildNodeSelectors, tr.Spec.PodTemplate.NodeSelector)
	})

	t.Run("Set build node selectors when only space node selectors is specified", func(t *testing.T) {
		testBuildNodeSelectors := make(map[string]string)

		testSpaceNodeSelectors := make(map[string]string)
		testSpaceNodeSelectors["abc"] = "abc"

		tr, err := MakeTaskRun(&v1alpha1.Build{},
			&tektonv1beta1.TaskSpec{},
			&v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{
						NodeSelector: testSpaceNodeSelectors,
					},
				},
				//Spec.RuntimeConfig.NodeSelector
			},
			&v1alpha1.SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-name",
					Namespace: "some-namespace",
				},
			},
			nil, &kfconfig.DefaultsConfig{
				BuildNodeSelectors: testBuildNodeSelectors,
			},
		)
		testutil.AssertNil(t, "err", err)
		testutil.AssertEqual(t, "buildNodeSelectors", testSpaceNodeSelectors, tr.Spec.PodTemplate.NodeSelector)
	})

	t.Run("Set build node selectors when both space node selectors and build node selectors are specified", func(t *testing.T) {
		testBuildNodeSelectors := make(map[string]string)
		testBuildNodeSelectors["disktype"] = "ssd"

		testSpaceNodeSelectors := make(map[string]string)
		testBuildNodeSelectors["abc"] = "abc"

		tr, err := MakeTaskRun(&v1alpha1.Build{},
			&tektonv1beta1.TaskSpec{},
			&v1alpha1.Space{
				Spec: v1alpha1.SpaceSpec{
					RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{
						NodeSelector: testSpaceNodeSelectors,
					},
				},
				//Spec.RuntimeConfig.NodeSelector
			},
			&v1alpha1.SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-name",
					Namespace: "some-namespace",
				},
			},
			nil, &kfconfig.DefaultsConfig{
				BuildNodeSelectors: testBuildNodeSelectors,
			},
		)
		testutil.AssertNil(t, "err", err)
		testutil.AssertEqual(t, "buildNodeSelectors", testBuildNodeSelectors, tr.Spec.PodTemplate.NodeSelector)
	})
}
