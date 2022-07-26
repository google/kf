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
	"context"
	"strings"
	"time"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/builds"
	"github.com/google/kf/v2/pkg/kf/dynamicutils"
	"github.com/google/kf/v2/pkg/kf/routes"
	clusterservicebrokers "github.com/google/kf/v2/pkg/kf/service-brokers/cluster"
	servicebrokers "github.com/google/kf/v2/pkg/kf/service-brokers/namespaced"
	"github.com/google/kf/v2/pkg/kf/serviceinstancebindings"
	"github.com/google/kf/v2/pkg/kf/serviceinstances"
	"github.com/google/kf/v2/pkg/kf/sourcepackages"
	"github.com/google/kf/v2/pkg/kf/spaces"
	"github.com/google/kf/v2/pkg/kf/tasks"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"knative.dev/pkg/apis"
)

func filterDeletionSet(obj *unstructured.Unstructured) bool {
	return obj.GetDeletionTimestamp() != nil
}

func filterGenerationStateDriftSet(obj *unstructured.Unstructured) bool {
	generation := obj.GetGeneration()
	observedGeneration, ok, _ := unstructured.NestedFieldNoCopy(obj.Object, "status", "observedGeneration")
	if !ok {
		return false
	}
	return observedGeneration != generation
}

// Status condition could be of type "Ready" or "Succeeded", return either one.
func getStatusCondition(obj *unstructured.Unstructured) map[string]interface{} {
	condition := getConditionByType(obj, string(apis.ConditionReady))
	if condition == nil {
		condition = getConditionByType(obj, string(apis.ConditionSucceeded))
	}
	return condition
}

// Determine if object's status condition (Ready|Succeeded) has False status.
func filterStatusFalseSet(obj *unstructured.Unstructured) bool {
	return dynamicutils.CheckCondtions(context.Background(), obj) == corev1.ConditionFalse
}

func filterReconliationErrorSet(obj *unstructured.Unstructured) bool {
	return hasStatusCondidtion(obj, "False", v1alpha1.ReconciliationError)
}

func filterTemplateErrorSet(obj *unstructured.Unstructured) bool {
	return hasStatusCondidtion(obj, "False", v1alpha1.TemplateError)
}

func filterChildNotOwnedSet(obj *unstructured.Unstructured) bool {
	return hasStatusCondidtion(obj, "False", v1alpha1.NotOwned)
}

func hasStatusCondidtion(obj *unstructured.Unstructured, targetStatus, targetReason string) bool {
	condition := getStatusCondition(obj)
	if condition == nil {
		return false
	}
	return matchConditionByStatusAndReason(condition, targetStatus, targetReason)
}

// Retrieve the condition from the object status by target condition type.
func getConditionByType(obj *unstructured.Unstructured, targetType string) map[string]interface{} {
	conditions, ok, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if ok {
		for _, condition := range conditions {
			conditionObj := condition.(map[string]interface{})
			conditionType, typeOk, _ := unstructured.NestedString(conditionObj, "type")
			if typeOk && conditionType == targetType {
				return conditionObj
			}
		}
	}
	return nil
}

// Given a condition, check if the condition's status and reason match the target values.
func matchConditionByStatusAndReason(condition map[string]interface{}, targetStatus, targetReason string) bool {
	status, statusOk, _ := unstructured.NestedString(condition, "status")
	if statusOk && status == targetStatus {
		reason, reasonOk, _ := unstructured.NestedString(condition, "reason")
		if reasonOk && reason == targetReason {
			return true
		}
	}
	return false
}

func filterBackingResourceReconcilationError(obj *unstructured.Unstructured) bool {
	condition := getConditionByType(obj, v1alpha1.BackingResourceReady)
	if condition != nil {
		return matchConditionByStatusAndReason(condition, "False", v1alpha1.ReconciliationError)
	}
	return false
}

func filterBackingResourceDeprovisionFailedError(obj *unstructured.Unstructured) bool {
	condition := getConditionByType(obj, v1alpha1.BackingResourceReady)
	if condition != nil {
		status, statusOk, _ := unstructured.NestedString(condition, "status")
		if statusOk && status == "False" {
			errorMessage, ok, _ := unstructured.NestedString(condition, "message")
			if ok {
				return strings.Contains(errorMessage, "DeprovisionFailed") || strings.Contains(errorMessage, "deprovision failed")
			}
		}
	}
	return false
}

func filterDeletionMayBeInFuture(obj *unstructured.Unstructured) bool {
	ts := obj.GetDeletionTimestamp()
	// Account for possible clock skew by subtracting a minute.
	return ts != nil && ts.After(time.Now().Add(-1*time.Minute))
}

func filterHasFinalizers(obj *unstructured.Unstructured) bool {
	return len(obj.GetFinalizers()) > 0
}

// objectStuckDeletingProblem diagnoses errors if an object is stuck deleting.
func objectStuckDeletingProblem() Problem {
	return Problem{
		Description: "Object is stuck deleting.",
		Filter:      filterDeletionSet,
		Causes: []Cause{
			{
				Description: "Deletion timestamp is in the future.",
				Filter:      filterDeletionMayBeInFuture,
				Recommendation: `With clock skew the ` + "`metadata.deletionTimestamp`" + ` may
        still be in the future. Wait a few minutes to see if the object is
        deleted.`,
			},

			{
				Description: "Finalizers exist on the object.",
				Filter:      filterHasFinalizers,
				Recommendation: `Finalizers are present on the object, they must be
        removed by the controller that set them before the object is deleted.

        If you want to force a deletion without waiting for the finalizers, edit
        the object to remove them from the ` + "`metadata.finalizers`" + ` array.

        To remove the finalizer from an object, use the
        ` + "`kubectl edit RESOURCE_TYPE RESOURCE_NAME -n my-space`" + ` command.

        See [using finalizers to control deletion](https://kubernetes.io/blog/2021/05/14/using-finalizers-to-control-deletion/) to learn more.

        Warning: Removing finalizers without allowing the controllers to complete
        may cause errors, security issues, data loss, or orphaned resources.`,
			},

			{
				Description: "Dependent objects may exist.",
				Filter:      nil,
				Recommendation: `The object may be waiting on dependents to be deleted before it is deleted.
        See the [Kubernetes garbage collection guide to learn more](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/).
        Have an administrator check all objects in the namespace and cluster to
        see if one of them is blocking deletion.

        If you need to remove the object without waiting for dependents, use
        ` + "`kubectl delete`" + ` with the cascade flag set to: ` + "`--cascade=orphan`" + `.`,
			},
		},
	}
}

// objectGenerationStateDriftProblem diagnoses errors if an object has drifted state cause by k8s object read eventual consistency.
func objectGenerationStateDriftProblem() Problem {
	return Problem{
		Description: "Object generation state drift.",
		Filter:      filterGenerationStateDriftSet,
		Causes: []Cause{
			{
				Description: "Object has generation version drift.",
				Filter:      filterGenerationStateDriftSet,
				Recommendation: `This error usually occurs Kf controller did not read the latest version of the object, this
		error is usually self-recovered once Kubernetes replicas reach eventual consistency, and it usually does not require
		action from users.`,
			},
		},
	}
}

// objectStatusFalseProblem diagnoses errors if an object has ready status false errors.
func objectStatusFalseProblem() Problem {
	return Problem{
		Description: "Object reconciliation failed.",
		Filter:      filterStatusFalseSet,
		Causes: []Cause{
			{
				Description: "Object has TemplateError",
				Filter:      filterTemplateErrorSet,
				Recommendation: `This error usually occurs if user has entered an invalid property in the custom resource
		Spec, or the configuration on the Space/Cluster is bad.

		To understand the root cause, user can read the longer error message in the object's ` + "`status.conditions`" + `
		using the command:` + "`kubectl describe RESOURCE_TYPE RESOURCE_NAME -n space`" + `. For example:
		` + "`kubectl describe serviceinstance my-service -n my-space`" + `.`,
			},
			{
				Description: "Object has ChildNotOwned error (Name conflicts)",
				Filter:      filterChildNotOwnedSet,
				Recommendation: `This error usually means that the object(s) the controller is trying to create already exists.
		This happens if the user created a K8s resource that has the same name as what the controller is trying to create;
		but more often it happens if user deletes a resource then Kf controller tries to re-create it. If a child resource
		is still hanging around, its owner will be the old resource that no longer exists.

		To recover from the error, it is recommended that user deletes the impacted resource and then recreates it. To delete the object,
		use a Kf deletion command or use the ` + "`kubectl delete RESOURCE_TYPE RESOURCE_NAME -n SPACE`" + `command. For example,
		` + "`kf delete-space my-space` or `kubectl delete space my-space`" + `.

		To recreate a resource, use a Kf command. For example: ` + "`kf create-space my-space`" + `.`,
			},
			{
				Description: "Object has ReconciliationError",
				Filter:      filterReconliationErrorSet,
				Recommendation: `This error usually means that something has gone wrong with the HTTP call made (by Kf controller)
		to the Kubernetes API servier to create/update resource.

		To understand the root cause, user can read the longer error message in the object's ` + "`status.conditions`" + `
		using the command:` + "`kubectl describe RESOURCE_TYPE RESOURCE_NAME -n space`" + `. For example:
		` + "`kubectl describe serviceinstance my-service -n my-space`" + `.`,
			},
		},
	}
}

// objectBackingResourceReconciliationProblem diagnoses errors if an object has backing resource reconciliation errors.
func objectBackingResourceReconciliationProblem() Problem {
	return Problem{
		Description: "Backing resource reconciliation failed.",
		Filter:      filterBackingResourceReconcilationError,
		Causes: []Cause{
			{
				Description: "Backing resource DeprovisionFailed error.",
				Filter:      filterBackingResourceDeprovisionFailedError,
				Recommendation: `This error usually occurs when backing resources (MySQL database hosted at an external OSB server)
		fails to be deprovisioned. Kf can not safely determine if the dependent resource is deprovisioned.

		To recover from the error, it is recommended that user reads the detail error message in the object's ` + "`status.conditions`" + `
		using the command:` + "`kubectl describe RESOURCE_TYPE RESOURCE_NAME -n space`" + `. For example:
		` + "`kubectl describe servipropogatesceinstance my-service -n my-space`" + `.` + `

		Once the error message is confirmed, have an administrator check the backing resource and clean it up manually. Once the backing
		resource is determined to be safely released, the impacted Kf resource can be reconciled successfully by manually
		removing the ` + "`Finalizer`" + ` from the object spec, use the ` + "`kubectl edit serviceinstance my-service -n my-space`" + `
		command.`,
			},
		},
	}
}

// CustomResourceComponents returns a list containing Kf's components.
func CustomResourceComponents() []Component {
	return []Component{
		AppComponent(),
		BuildComponent(),
		ClusterServiceBrokerComponent(),
		RouteComponent(),
		ServiceBrokerComponent(),
		ServiceInstanceComponent(),
		ServiceInstanceBindingComponent(),
		SourcePackageComponent(),
		SpaceComponent(),
		TaskComponent(),
	}
}

// AppComponent specifies troubleshooting for a Kf App.
func AppComponent() Component {
	return Component{
		Type: apps.NewResourceInfo(),
		Problems: []Problem{
			objectStuckDeletingProblem(),
			objectGenerationStateDriftProblem(),
			objectStatusFalseProblem(),
		},
	}
}

// BuildComponent specifies troubleshooting for a Kf Build.
func BuildComponent() Component {
	return Component{
		Type: builds.NewResourceInfo(),
		Problems: []Problem{
			objectStuckDeletingProblem(),
			objectGenerationStateDriftProblem(),
			objectStatusFalseProblem(),
		},
	}
}

// ClusterServiceBrokerComponent specifies troubleshooting for a Kf ClusterServiceBroker.
func ClusterServiceBrokerComponent() Component {
	return Component{
		Type: clusterservicebrokers.NewResourceInfo(),
		Problems: []Problem{
			objectStuckDeletingProblem(),
			objectGenerationStateDriftProblem(),
			objectStatusFalseProblem(),
		},
	}
}

// RouteComponent specifies troubleshooting for a Kf Route.
func RouteComponent() Component {
	return Component{
		Type: routes.NewResourceInfo(),
		Problems: []Problem{
			objectStuckDeletingProblem(),
			objectGenerationStateDriftProblem(),
			objectStatusFalseProblem(),
		},
	}
}

// ServiceBrokerComponent specifies troubleshooting for a Kf ServiceBroker.
func ServiceBrokerComponent() Component {
	return Component{
		Type: servicebrokers.NewResourceInfo(),
		Problems: []Problem{
			objectStuckDeletingProblem(),
			objectGenerationStateDriftProblem(),
			objectStatusFalseProblem(),
		},
	}
}

// ServiceInstanceComponent specifies troubleshooting for a Kf Service.
func ServiceInstanceComponent() Component {
	return Component{
		Type: serviceinstances.NewResourceInfo(),
		Problems: []Problem{
			objectStuckDeletingProblem(),
			objectGenerationStateDriftProblem(),
			objectStatusFalseProblem(),
			objectBackingResourceReconciliationProblem(),
		},
	}
}

// ServiceInstanceBindingComponent specifies troubleshooting for a Kf Binding.
func ServiceInstanceBindingComponent() Component {
	return Component{
		Type: serviceinstancebindings.NewResourceInfo(),
		Problems: []Problem{
			objectStuckDeletingProblem(),
			objectGenerationStateDriftProblem(),
			objectStatusFalseProblem(),
		},
	}
}

// SourcePackageComponent specifies troubleshooting for a Kf SourcePackage.
func SourcePackageComponent() Component {
	return Component{
		Type: sourcepackages.NewResourceInfo(),
		Problems: []Problem{
			objectStuckDeletingProblem(),
			objectGenerationStateDriftProblem(),
			objectStatusFalseProblem(),
		},
	}
}

// SpaceComponent specifies troubleshooting for a Kf Space.
func SpaceComponent() Component {
	return Component{
		Type: spaces.NewResourceInfo(),
		Problems: []Problem{
			objectStuckDeletingProblem(),
			objectGenerationStateDriftProblem(),
			objectStatusFalseProblem(),
		},
	}
}

// TaskComponent specifies troubleshooting for a Kf Task.
func TaskComponent() Component {
	return Component{
		Type: tasks.NewResourceInfo(),
		Problems: []Problem{
			objectStuckDeletingProblem(),
			objectGenerationStateDriftProblem(),
			objectStatusFalseProblem(),
		},
	}
}
