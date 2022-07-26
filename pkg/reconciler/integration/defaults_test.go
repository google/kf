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
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

// DefaultTest defines validation for defaulting webhooks. The Resource will
// be created and all paths will be tested for equality.
type DefaultTest struct {
	Path              string                 `json:"-"`
	TestEqualityPaths []string               `json:"testEquality"`
	Resource          map[string]interface{} `json:"resource"`
}

func (dt *DefaultTest) GetUnstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: dt.Resource,
	}
}

func (dt *DefaultTest) GetGVR() schema.GroupVersionResource {
	obj := dt.GetUnstructured()
	gvk := obj.GroupVersionKind()

	return schema.GroupVersionResource{
		Group:   gvk.Group,
		Version: gvk.Version,

		// We can autodiscover this in the future from the API Server like
		// Kubernetes does, but in the meantime all resources we deal with follow
		// this formula.
		Resource: strings.ToLower(gvk.Kind) + "s",
	}
}

func loadTests(t *testing.T) []*DefaultTest {
	t.Helper()

	tests, err := filepath.Glob(filepath.Join("testdata", "defaults", "*.yaml"))
	testutil.AssertNil(t, "defaults err", err)

	// Sort so the tests are deterministic as different filesystems may traverse
	// their directories in different ways.
	sort.Strings(tests)

	var out []*DefaultTest

	for _, testPath := range tests {
		testBytes, err := ioutil.ReadFile(testPath)
		testutil.AssertNil(t, "read error", err)

		defaultTest := &DefaultTest{}
		err = yaml.Unmarshal(testBytes, defaultTest)
		testutil.AssertNil(t, "unmarshal error", err)
		defaultTest.Path = testPath

		out = append(out, defaultTest)
	}

	return out
}

// TestLoadTests validates that tests can be loaded from disk
// this test will run even if integration tests are not enabled.
func TestLoadTests(t *testing.T) {
	loadTests(t)
}

func TestIntegration_ApplyDefaults(t *testing.T) {
	t.Skip("b/187452966")

	for _, tc := range loadTests(t) {
		// Shadow to avoid closure issues.
		tc := tc

		integration.RunKubeAPITest(context.Background(), t, func(ctx context.Context, t *testing.T) {
			t.Run(tc.Path, func(t *testing.T) {
				integration.WithSpace(ctx, t, func(namespace string) {
					integration.WithDynamicClient(ctx, t, func(client dynamic.Interface) {

						dyn := client.Resource(tc.GetGVR()).Namespace(namespace)

						obj := tc.GetUnstructured()
						obj.SetNamespace(namespace)
						expected := obj.DeepCopy()

						actual, err := dyn.Create(context.Background(), obj, metav1.CreateOptions{})
						testutil.AssertNil(t, "creation error", err)

						// Individual paths are tested because Kubernetes doesn't have a
						// strict system about what is/isn't allowed to be updated
						// or the structure of resources.
						// We need to ignore things we know are going to be set by the
						// controller like the status, metadata UUID, etc. But maybe also
						// things in the spec. For example, types that increment a
						// generation field.
						for _, path := range tc.TestEqualityPaths {
							t.Run(path, func(t *testing.T) {
								AssertPathEqual(t, path, expected, actual)
							})
						}
					})
				})
			})
		})
	}
}

func AssertPathEqual(t *testing.T, path string, expected, actual *unstructured.Unstructured) {
	t.Helper()

	path = strings.TrimPrefix(path, ".")
	splitPath := strings.Split(path, ".")

	ev, eok, eerr := unstructured.NestedFieldNoCopy(expected.UnstructuredContent(), splitPath...)
	testutil.AssertNil(t, "expected path error", eerr)

	av, aok, aerr := unstructured.NestedFieldNoCopy(actual.UnstructuredContent(), splitPath...)
	testutil.AssertNil(t, "actual path error", aerr)

	// Test ok before trying to marshal objects that might be bad.
	testutil.AssertEqual(t, "ok", eok, aok)

	// Convert the objects to JSON, this is so we can ignore weird encoding issues
	// like int32 vs int64.
	evJSON, eerr := json.Marshal(ev)
	testutil.AssertNil(t, "expected JSON error", eerr)

	avJSON, aerr := json.Marshal(av)
	testutil.AssertNil(t, "actual JSON error", aerr)

	if !reflect.DeepEqual(evJSON, avJSON) {
		testutil.AssertEqual(t, "values", ev, av)
	}
}
