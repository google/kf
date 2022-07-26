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

package troubleshooter

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func MetaTime(t time.Time) *metav1.Time {
	return &metav1.Time{
		Time: t,
	}
}

func ExampleTroubleshootingTest_filterDeletionSet() {
	obj := &unstructured.Unstructured{}
	fmt.Println("Not deleted:", filterDeletionSet(obj))

	obj.SetDeletionTimestamp(MetaTime(time.Now()))
	fmt.Println("Deleted:", filterDeletionSet(obj))

	// Output: Not deleted: false
	// Deleted: true
}

func ExampleTroubleshootingTest_filterGenerationStateDriftSet() {
	obj1 := &unstructured.Unstructured{}

	obj1.Object = map[string]interface{}{
		"status": map[string]interface{}{
			"observedGeneration": obj1.GetGeneration(),
		},
	}

	obj2 := &unstructured.Unstructured{}

	obj2.Object = map[string]interface{}{
		"status": map[string]interface{}{
			"observedGeneration": obj2.GetGeneration() + 1,
		},
	}

	fmt.Println("Obj1 Generation drifted:", filterGenerationStateDriftSet(obj1))
	fmt.Println("Obj2 Generation drifted:", filterGenerationStateDriftSet(obj2))

	// Output: Obj1 Generation drifted: false
	// Obj2 Generation drifted: true
}

func ExampleTroubleshootingTest_filterStatusFalseSet() {
	obj1 := &unstructured.Unstructured{}
	obj1.Object = map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Ready",
					"status": "False",
				},
			},
		},
	}

	obj2 := &unstructured.Unstructured{}
	obj2.Object = map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Succeeded",
					"status": "False",
				},
			},
		},
	}

	fmt.Println("Obj1 is matched:", filterStatusFalseSet(obj1))
	fmt.Println("Obj2 is matched:", filterStatusFalseSet(obj2))

	// Output: Obj1 is matched: true
	// Obj2 is matched: true
}

func ExampleTroubleshootingTest_filterReconliationErrorSet() {
	obj := &unstructured.Unstructured{}
	obj.Object = map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Ready",
					"status": "False",
					"reason": "ReconciliationError",
				},
			},
		},
	}

	fmt.Println("Obj has ReconciliationError:", filterReconliationErrorSet(obj))
	// Output: Obj has ReconciliationError: true
}

func ExampleTroubleshootingTest_filterTemplateErrorSet() {
	obj := &unstructured.Unstructured{}
	obj.Object = map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Ready",
					"status": "False",
					"reason": "TemplateError",
				},
			},
		},
	}

	fmt.Println("Obj has TemplateError:", filterTemplateErrorSet(obj))
	// Output: Obj has TemplateError: true
}

func ExampleTroubleshootingTest_filterChildNotOwnedSet() {
	obj := &unstructured.Unstructured{}
	obj.Object = map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Ready",
					"status": "False",
					"reason": "NotOwned",
				},
			},
		},
	}

	fmt.Println("Obj has NotOwned error:", filterChildNotOwnedSet(obj))
	// Output: Obj has NotOwned error: true
}

func ExampleTroubleshootingTest_filterBackingResourceReconcilationError() {
	obj := &unstructured.Unstructured{}
	obj.Object = map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "BackingResourceReady",
					"status": "False",
					"reason": "ReconciliationError",
				},
			},
		},
	}

	fmt.Println("Obj has BackingResourceReady error:", filterBackingResourceReconcilationError(obj))
	// Output: Obj has BackingResourceReady error: true
}

func ExampleTroubleshootingTest_filterBackingResourceDeprovisionFailedError() {
	obj1 := &unstructured.Unstructured{}
	obj1.Object = map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":    "BackingResourceReady",
					"status":  "False",
					"message": "backing resource DeprovisionFailed",
				},
			},
		},
	}

	obj2 := &unstructured.Unstructured{}
	obj2.Object = map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":    "BackingResourceReady",
					"status":  "False",
					"message": "backing resource deprovision failed",
				},
			},
		},
	}
	fmt.Println("Obj1 has DeprovisionFailed error:", filterBackingResourceDeprovisionFailedError(obj1))
	fmt.Println("Obj2 has DeprovisionFailed error:", filterBackingResourceDeprovisionFailedError(obj2))
	// Output: Obj1 has DeprovisionFailed error: true
	// Obj2 has DeprovisionFailed error: true
}

func ExampleTroubleshootingTest_filterDeletionMayBeInFuture() {
	obj := &unstructured.Unstructured{}

	obj.SetDeletionTimestamp(MetaTime(time.Now()))
	fmt.Println("Deleted recently:", filterDeletionMayBeInFuture(obj))

	obj.SetDeletionTimestamp(MetaTime(time.Now().Add(-2 * time.Minute)))
	fmt.Println("Deleted in past:", filterDeletionMayBeInFuture(obj))

	// Output: Deleted recently: true
	// Deleted in past: false
}

func ExampleTroubleshootingTest_filterHasFinalizers() {
	obj := &unstructured.Unstructured{}

	fmt.Println("No finalizers:", filterHasFinalizers(obj))

	obj.SetFinalizers([]string{"foo", "bar"})
	fmt.Println("With finalizers:", filterHasFinalizers(obj))

	// Output: No finalizers: false
	// With finalizers: true
}

func assertCustomResourceInvariants(t *testing.T, c Component) {

	t.Run("checks deletion first", func(t *testing.T) {
		if len(c.Problems) == 0 {
			t.Fatal("expected at least one Problems")
		}

		testutil.AssertEqual(
			t,
			"first problem checked should be deletion",
			objectStuckDeletingProblem().Description,
			c.Problems[0].Description,
		)
	})

	// Validate properties of the objects.
	for i, problem := range c.Problems {
		t.Run(fmt.Sprintf("Problems[%d]", i), func(t *testing.T) {
			testutil.AssertTrue(t, "Description not blank", problem.Description != "")
			testutil.AssertNotNil(t, "Filter not nil", problem.Filter)

			for j, cause := range problem.Causes {
				t.Run(fmt.Sprintf("Causes[%d]", j), func(t *testing.T) {
					testutil.AssertTrue(t, "Description not blank", cause.Description != "")

					// Filters can be nil or not so don't validate.

					testutil.AssertTrue(t, "Recommendation not blank", cause.Description != "")
				})
			}
		})
	}
}

func TestAppComponent(t *testing.T) {
	component := AppComponent()
	assertCustomResourceInvariants(t, component)
}

func TestBuildComponent(t *testing.T) {
	component := BuildComponent()
	assertCustomResourceInvariants(t, component)
}

func TestClusterServiceBrokerComponent(t *testing.T) {
	component := ClusterServiceBrokerComponent()
	assertCustomResourceInvariants(t, component)
}

func TestRouteComponent(t *testing.T) {
	component := RouteComponent()
	assertCustomResourceInvariants(t, component)
}

func TestServiceBrokerComponent(t *testing.T) {
	component := ServiceBrokerComponent()
	assertCustomResourceInvariants(t, component)
}

func TestServiceInstanceComponent(t *testing.T) {
	component := ServiceInstanceComponent()
	assertCustomResourceInvariants(t, component)
}

func TestServiceInstanceBindingComponent(t *testing.T) {
	component := ServiceInstanceBindingComponent()
	assertCustomResourceInvariants(t, component)
}

func TestSourcePackageComponent(t *testing.T) {
	component := SourcePackageComponent()
	assertCustomResourceInvariants(t, component)
}

func TestSpaceComponent(t *testing.T) {
	component := SpaceComponent()
	assertCustomResourceInvariants(t, component)
}

func TestTaskComponent(t *testing.T) {
	component := TaskComponent()
	assertCustomResourceInvariants(t, component)
}
