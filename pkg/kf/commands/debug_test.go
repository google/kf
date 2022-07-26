// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"context"
	"os"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	fakek8s "k8s.io/client-go/kubernetes/fake"
)

func mockPodTemplateSpec(images ...string) (out corev1.PodTemplateSpec) {
	for _, image := range images {
		out.Spec.Containers = append(out.Spec.Containers, corev1.Container{
			Image: image,
		})
	}

	return
}

func TestGatherContainerImages(t *testing.T) {
	cases := map[string]struct {
		template   corev1.PodTemplateSpec
		wantImages []string
	}{
		"blank": {
			template:   mockPodTemplateSpec(),
			wantImages: []string{},
		},

		"single image": {
			template:   mockPodTemplateSpec("gcr.io/test"),
			wantImages: []string{"gcr.io/test"},
		},

		"multiple images": {
			template:   mockPodTemplateSpec("gcr.io/c", "gcr.io/a", "gcr.io/b"),
			wantImages: []string{"gcr.io/a", "gcr.io/b", "gcr.io/c"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gotImages := gatherContainerImages(tc.template)
			testutil.AssertEqual(t, "images", tc.wantImages, gotImages)
		})
	}
}

func mockDeployment(namespace, name string, images ...string) *appsv1.Deployment {
	out := &appsv1.Deployment{}
	out.Namespace = namespace
	out.Name = name
	out.Spec.Template = mockPodTemplateSpec(images...)

	return out
}

func mockDaemonSet(namespace, name string, images ...string) *appsv1.DaemonSet {
	out := &appsv1.DaemonSet{}
	out.Namespace = namespace
	out.Name = name
	out.Spec.Template = mockPodTemplateSpec(images...)

	return out
}

func ExampledebugServerComponents() {
	client := fakek8s.NewSimpleClientset(
		mockDeployment("kf", "webhook", "gcr.io/kf-releases/webhook:v1.0.0"),
		mockDaemonSet("kube-system", "wi", "gke.io/wi/wi:v1.0.0"),
	)

	debugServerComponents(context.Background(), os.Stdout, client)

	// Output: Cluster Components:
	//   Namespace    Resource            Images
	//   kf           Deployment/webhook  [gcr.io/kf-releases/webhook:v1.0.0]
	//   kube-system  DaemonSet/wi        [gke.io/wi/wi:v1.0.0]
}
