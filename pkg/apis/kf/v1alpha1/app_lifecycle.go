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

package v1alpha1

import (
	"fmt"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (r *App) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("App")
}

// PropagateBuildStatus copies the Build status to the App's.
func (status *AppStatus) PropagateBuildStatus(build *Build) {
	status.LatestCreatedBuildName = build.Name

	cond := build.Status.GetCondition(BuildConditionSucceeded)
	if PropagateCondition(status.manage(), AppConditionBuildReady, cond) {
		status.LatestReadyBuildName = build.Name
		status.BuildStatusFields = build.Status.BuildStatusFields
		status.Image = build.Status.Image
	}
}

// PropagateADXBuildStatus copies the AppDevExperience Build status to the
// App's.
func (status *AppStatus) PropagateADXBuildStatus(u *unstructured.Unstructured) error {
	status.LatestCreatedBuildName = u.GetName()
	status.LatestReadyBuildName = u.GetName()

	image, _, err := unstructured.NestedString(u.Object, "status", "image")
	if err != nil {
		return fmt.Errorf("failed to read image from status: %v", err)
	}

	status.Image = image
	return nil
}

// PropagateInstanceStatus updates the effective instance status from the app
func (status *AppStatus) PropagateInstanceStatus(is InstanceStatus) {
	status.Instances = is
}

// PropagateAutoscalingStatus updates the effective instance status with autoscaling status.
func (status *InstanceStatus) PropagateAutoscalingStatus(app *App, hpa *autoscalingv1.HorizontalPodAutoscaler) {
	// hpa is nil when autoscaling is disabled, or maxreplicas hasn't been set, or no rules specified by user.
	if hpa == nil {
		return
	}

	if hpa.Status.CurrentCPUUtilizationPercentage == nil {
		// hpa is not ready yet
		return
	}

	// Set instance status for autoscaling rule
	status.AutoscalingStatus = []AutoscalingRuleStatus{
		{
			// Rules is guaranteed to not be empty here because hpa is not nil.
			AppAutoscalingRule: app.Spec.Instances.Autoscaling.Rules[0],
			Current: AutoscalingRuleMetricValueStatus{
				AverageValue: resource.NewQuantity(int64(*hpa.Status.CurrentCPUUtilizationPercentage), resource.DecimalSI),
			},
		},
	}
}

// PropagateAutoscalerV1Status updates the autoscaler status to reflect the
// underlying state of the autoscaler.
// HorizontalPodAutoscalerCondition is not available for V1.
func (status *AppStatus) PropagateAutoscalerV1Status(autoscaler *autoscalingv1.HorizontalPodAutoscaler) {

	if autoscaler == nil {
		status.HorizontalPodAutoscalerCondition().MarkSuccess()
		return
	}

	if autoscaler.Status.CurrentReplicas > autoscaler.Status.DesiredReplicas {
		status.HorizontalPodAutoscalerCondition().MarkUnknown("ScalingDown", "waiting for autoscaler to finish scaling down: current replicas %d, target replicas %d", autoscaler.Status.CurrentReplicas, autoscaler.Status.DesiredReplicas)
		return
	}

	if autoscaler.Status.CurrentReplicas < autoscaler.Status.DesiredReplicas {
		status.HorizontalPodAutoscalerCondition().MarkUnknown("ScalingUp", "waiting for autoscaler to finish scaling up: current replicas %d, target replicas %d", autoscaler.Status.CurrentReplicas, autoscaler.Status.DesiredReplicas)
		return
	}

	status.HorizontalPodAutoscalerCondition().MarkSuccess()
}

// PropagateDeploymentStatus updates the deployment status to reflect the
// underlying state of the deployment.
func (status *AppStatus) PropagateDeploymentStatus(deployment *appsv1.Deployment) {

	for _, cond := range deployment.Status.Conditions {
		// ReplicaFailure is added in a deployment when one of its pods fails to be created
		// or deleted.
		if cond.Type == appsv1.DeploymentReplicaFailure && cond.Status == corev1.ConditionTrue {
			status.DeploymentCondition().MarkFalse(cond.Reason, cond.Message)
			return
		}
	}

	if deployment.Generation > deployment.Status.ObservedGeneration {
		status.DeploymentCondition().MarkUnknown("GenerationOutOfDate", "waiting for deployment spec update to be observed")
		return
	}

	for _, cond := range deployment.Status.Conditions {
		if cond.Type == appsv1.DeploymentProgressing && cond.Reason == "ProgressDeadlineExceeded" {
			status.DeploymentCondition().MarkFalse("DeadlineExceeded", "deployment %q exceeded its progress deadline", deployment.Name)
			return
		}
	}

	if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
		status.DeploymentCondition().MarkUnknown("UpdatingReplicas", "waiting for deployment %q rollout to finish: %d out of %d new replicas have been updated", deployment.Name, deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas)
		return
	}
	if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
		status.DeploymentCondition().MarkUnknown("TerminatingOldReplicas", "waiting for deployment %q rollout to finish: %d old replicas are pending termination", deployment.Name, deployment.Status.Replicas-deployment.Status.UpdatedReplicas)
		return
	}
	if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
		status.DeploymentCondition().MarkUnknown("InitializingPods", "waiting for deployment %q rollout to finish: %d of %d updated replicas are available", deployment.Name, deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas)
		return
	}

	status.DeploymentCondition().MarkSuccess()
}

// PropagateEnvVarSecretStatus updates the env var secret readiness status.
func (status *AppStatus) PropagateEnvVarSecretStatus(secret *v1.Secret) {
	status.EnvVarSecretCondition().MarkSuccess()
}

// PropagateServiceStatus propagates the app's internal URL.
func (status *AppStatus) PropagateServiceStatus(service *corev1.Service) {
	// Services don't have conditions to indicate readiness.
	status.ServiceCondition().MarkSuccess()
}

// PropagateServiceAccountStatus propagates the app's service account name.
func (status *AppStatus) PropagateServiceAccountStatus(serviceAccount *corev1.ServiceAccount) {
	status.ServiceAccountName = serviceAccount.Name
	status.ServiceAccountCondition().MarkSuccess()
}

// PropagateRouteStatus updates the route readiness status.
func (status *AppStatus) PropagateRouteStatus(bindings []QualifiedRouteBinding, routes []Route, undeclaredBindings []QualifiedRouteBinding) {
	var rs []AppRouteStatus
	var conditions []apis.Condition
	var urls []string

	// Ensure output is deterministic
	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].Source.String() < bindings[j].Source.String()
	})

	sort.Slice(undeclaredBindings, func(i, j int) bool {
		return undeclaredBindings[i].Source.String() < undeclaredBindings[j].Source.String()
	})

	// Create a mapping for lookup later
	routeMapping := make(map[RouteSpecFields]Route)
	for _, route := range routes {
		routeMapping[route.Spec.RouteSpecFields] = route
	}

	for idx, binding := range bindings {
		cond := apis.Condition{
			Type: apis.ConditionType(fmt.Sprintf("Route%dReady", idx)),
		}

		bindingStatus := AppRouteStatus{
			QualifiedRouteBinding: binding,
			URL:                   binding.Source.String(),
			Status:                RouteBindingStatusUnknown,
		}

		route, routeFound := routeMapping[binding.Source]

		// Reasons all start with Route so when they're propagated all the way
		// up to the App's status it's easy to determine what went wrong.
		switch {
		case !routeFound:
			cond.Status = corev1.ConditionFalse
			cond.Reason = "RouteMissing"
			cond.Message = fmt.Sprintf("No Route defined for URL: %s", bindingStatus.URL)

		case route.Generation != route.Status.ObservedGeneration:
			cond.Status = corev1.ConditionUnknown
			cond.Reason = "RouteReconciling"
			cond.Message = "The Route is currently updating"

		case !route.Status.IsReady():
			rc := route.Status.GetCondition(RouteConditionReady)
			cond.Status = rc.Status
			cond.Reason = "RouteUnhealthy"
			cond.Message = fmt.Sprintf("Route has status %s: %s", rc.Reason, rc.Message)

		case !route.hasDestination(binding.Destination):
			cond.Status = corev1.ConditionUnknown
			cond.Reason = "RouteBindingPropagating"
			cond.Message = "The binding is still propagating to the Route"

		default:
			cond.Status = corev1.ConditionTrue
			cond.Reason = "RouteReady"
			cond.Message = "The Route is up to date and is mapped to this App"

			bindingStatus.Status = RouteBindingStatusReady
			bindingStatus.VirtualService = route.Status.VirtualService
		}

		conditions = append(conditions, cond)
		rs = append(rs, bindingStatus)
		urls = append(urls, bindingStatus.URL)
	}

	// Add statuses for all bindings that exist but aren't on the App.
	// This usually happens as the result of deleting an binding from
	// an App, the result should be reconciled away relatively quickly.
	// Bindings are appended with a deleting status to prevent the
	// route reconciler from getting stuck in a catch-22 but still allow
	// it to enqueue the routes that are holding old bindings.
	for _, binding := range undeclaredBindings {
		rs = append(rs, AppRouteStatus{
			QualifiedRouteBinding: binding,
			URL:                   binding.Source.String(),
			Status:                RouteBindingStatusOrphaned,
		})

		conditions = append(conditions, apis.Condition{
			Type:    apis.ConditionType(fmt.Sprintf("Route%dReady", len(rs))),
			Status:  corev1.ConditionUnknown,
			Reason:  "ExtraRouteBinding",
			Message: fmt.Sprintf("The Route %s has an extra binding to this App", binding.Source.String()),
		})
		urls = append(urls, binding.Source.String())
	}

	// Sort the URLs for deterministic output.
	sort.Strings(urls)
	status.URLs = urls

	summaryCondition, allConditions := SummarizeChildConditions(conditions)

	status.Routes = rs
	PropagateCondition(
		status.manage(),
		AppConditionRouteReady,
		summaryCondition,
	)
	status.RouteConditions = allConditions
}

// PropagateVolumeBindingsStatus updates the service binding readiness status.
func (status *AppStatus) PropagateVolumeBindingsStatus(volumeBindings []*ServiceInstanceBinding) {
	volumeStatus := []AppVolumeStatus{}

	for _, binding := range volumeBindings {
		volumeStatus = append(volumeStatus, AppVolumeStatus{
			VolumeName:      binding.Status.VolumeStatus.PersistentVolumeName,
			MountPath:       binding.Status.VolumeStatus.Mount,
			VolumeClaimName: binding.Status.VolumeStatus.PersistentVolumeClaimName,
			ReadOnly:        binding.Status.VolumeStatus.ReadOnly,
			UidGid: UidGid{
				GID: binding.Status.VolumeStatus.GID,
				UID: binding.Status.VolumeStatus.UID,
			},
		})
	}

	// Make sure status is deterministic.
	sort.Slice(volumeStatus, func(i, j int) bool {
		return volumeStatus[i].MountPath < volumeStatus[j].MountPath
	})
	status.Volumes = volumeStatus
}

// PropagateServiceInstanceBindingsStatus updates the service binding readiness status.
func (status *AppStatus) PropagateServiceInstanceBindingsStatus(bindings []ServiceInstanceBinding) {
	// Make sure binding sorting is deterministic.
	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].Name < bindings[j].Name
	})

	// Gather binding names
	var bindingNames []string
	for _, binding := range bindings {
		bindingNames = append(bindingNames, binding.Status.BindingName)
	}
	status.ServiceBindingNames = bindingNames

	// Gather binding conditions
	var conditionTypes []apis.ConditionType
	for _, binding := range bindings {
		conditionType := serviceBindingConditionType(binding)
		conditionTypes = append(conditionTypes, conditionType)
	}

	duckStatus := &duckv1beta1.Status{}
	manager := apis.NewLivingConditionSet(conditionTypes...).Manage(duckStatus)
	manager.InitializeConditions()

	for _, binding := range bindings {
		if binding.Generation != binding.Status.ObservedGeneration {
			// this binding's conditions are out of date.
			continue
		}
		for _, cond := range binding.Status.Conditions {
			if cond.Type != ServiceInstanceBindingConditionReady {
				continue
			}

			conditionType := serviceBindingConditionType(binding)
			switch v1.ConditionStatus(cond.Status) {
			case v1.ConditionTrue:
				manager.MarkTrue(conditionType)
			case v1.ConditionFalse:
				manager.MarkFalse(conditionType, cond.Reason, cond.Message)
			case v1.ConditionUnknown:
				manager.MarkUnknown(conditionType, cond.Reason, cond.Message)
			}
		}
	}

	// if there are no bindings, set the happy condition to true
	if len(bindings) == 0 {
		manager.MarkTrue(apis.ConditionReady)
	}

	// Copy Ready condition
	PropagateCondition(status.manage(), AppConditionServiceInstanceBindingsReady, manager.GetCondition(apis.ConditionReady))
	status.ServiceBindingConditions = duckStatus.Conditions
}

// MarkSpaceHealthy notes that the space was able to be retrieved and
// defaults can be applied from it.
func (status *AppStatus) MarkSpaceHealthy() {
	status.SpaceCondition().MarkSuccess()
}

// MarkSpaceUnhealthy notes that the space was could not be retrieved.
func (status *AppStatus) MarkSpaceUnhealthy(reason, message string) {
	status.SpaceCondition().MarkFalse(reason, message)
}

// serviceBindingConditionType creates a Conditiontype for a ServiceBinding.
func serviceBindingConditionType(binding ServiceInstanceBinding) apis.ConditionType {
	serviceBinding := binding.Status.BindingName
	return apis.ConditionType(fmt.Sprintf("%sReady", serviceBinding))
}
