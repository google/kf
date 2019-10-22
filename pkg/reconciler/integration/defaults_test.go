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

package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/kf/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

func TestIntegration_ApplyDefaults(t *testing.T) {
	cases := map[string]struct {
		Group    string
		Version  string
		Resource string
		File     string
	}{
		"serving": {
			Group:    "serving.knative.dev",
			Version:  "v1alpha1",
			Resource: "services",
			File:     "serving_v1alpha1_defaults.yaml",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			testutil.RunKubeAPITest(t, func(ctx context.Context, t *testing.T) {
				testutil.WithNamespace(ctx, t, func(namespace string) {
					testutil.WithDynamicClient(ctx, t, func(client dynamic.Interface) {
						// pass
						dyn := client.Resource(schema.GroupVersionResource{
							Group:    tc.Group,
							Version:  tc.Version,
							Resource: tc.Resource,
						}).Namespace(namespace)

						objPath := filepath.Join("testdata", tc.File)
						objBytes, err := ioutil.ReadFile(objPath)
						testutil.AssertNil(t, "read error", err)

						unstructuredContent := make(map[string]interface{})
						err = yaml.Unmarshal(objBytes, &unstructuredContent)
						testutil.AssertNil(t, "unmarshal error", err)

						obj := &unstructured.Unstructured{}
						obj.SetUnstructuredContent(unstructuredContent)
						obj.SetNamespace(namespace)

						out, err := dyn.Create(obj, metav1.CreateOptions{})
						testutil.AssertNil(t, "creation error", err)
						fmt.Println(out.UnstructuredContent())
					})
				})
			})
		})
	}

}
