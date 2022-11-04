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

package v1alpha1

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	autoscaling "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const (
	// NameLabel holds the standard label key for Kubernetes app names.
	NameLabel = "app.kubernetes.io/name"
	// ManagedByLabel holds the standard label key for Kubernetes app managers.
	ManagedByLabel = "app.kubernetes.io/managed-by"
	// ComponentLabel holds the standard label key for Kubernetes app component
	// types.
	ComponentLabel = "app.kubernetes.io/component"
	// VersionLabel holds the current version for a Kubernetes app.
	VersionLabel = "app.kubernetes.io/version"
	// FeatureFlagsAnnotation holds a map of each feature flag to a bool indicating whether the feature is enabled or not.
	FeatureFlagsAnnotation = "kf.dev/feature-flags"
	// WorkloadIdentityAnnotation is the annotation used to map Kubernetes
	// Service Accounts (KSA) to a Google Service Accounts (GSA).
	WorkloadIdentityAnnotation = "iam.gke.io/gcp-service-account"
	// DefaultUserContainerName contains the default name for the user container.
	DefaultUserContainerName = "user-container"
	// DefaultMaxTaskCount is the maximum number of tasks to keep in an App.
	DefaultMaxTaskCount = 500
	// AppServerComponent is the value used for the App component.
	AppServerComponent = "app-server"
)

// RouteBindingStatus represents the status of a RouteBinding.
type RouteBindingStatus string

const (
	// RouteBindingStatusOrphaned indicates the Binding isn't desired on the App but is
	// still waiting to be reconciled off the VirtualService.
	RouteBindingStatusOrphaned RouteBindingStatus = "Orphaned"

	// RouteBindingStatusUnknown indicates the status is unknown at this time.
	RouteBindingStatusUnknown RouteBindingStatus = "Unknown"

	// RouteBindingStatusReady indicates the binding is ready to receive traffic.
	RouteBindingStatusReady RouteBindingStatus = "Ready"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// App is a 12-factor application deployed to Knative. It encompasses source
// code, configuration, and the current state of the application.
type App struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec AppSpec `json:"spec,omitempty"`

	// +optional
	Status AppStatus `json:"status,omitempty"`
}

var _ apis.Validatable = (*App)(nil)
var _ apis.Defaultable = (*App)(nil)

// ComponentLabels returns Kubernetes recommended labels to tie together
// deployed applications and their pieces. The provided component name
// specifies the sub-resource of the app e.g. "database", "load-balancer",
// or "server".
func (app *App) ComponentLabels(component string) map[string]string {
	return AppComponentLabels(app.Name, component)
}

// LogSelector implements logs.Object.
func (app *App) LogSelector(children bool) string {
	labels := AppComponentLabels(app.Name, AppServerComponent)

	if children {
		// We want EVERYTHING that the App owns (e.g., Build pods).
		delete(labels, ComponentLabel)
	}

	return metav1.FormatLabelSelector(metav1.SetAsLabelSelector(labels))
}

// AppComponentLabels creates the labels matching the given component and app.
func AppComponentLabels(appName, component string) map[string]string {
	return map[string]string{
		NameLabel:      appName,
		ManagedByLabel: "kf",
		ComponentLabel: component,
	}
}

// AppSpec is the desired configuration for an App.
type AppSpec struct {

	// Build defines the App's build configuration.
	Build AppSpecBuild `json:"build"`

	// Template defines the App's runtime configuration.
	// +optional
	Template AppSpecTemplate `json:"template"`

	// Instances defines the scaling rules for the App.
	Instances AppSpecInstances `json:"instances,omitempty"`

	// Routes defines the routing rules for the App.
	// +optional
	// +patchStrategy=merge
	Routes []RouteWeightBinding `json:"routes,omitempty"`
}

// AppSpecBuild defines an app's build configuration.
type AppSpecBuild struct {

	// UpdateRequests is a unique identifier for a Build.
	// Updating sub-values will trigger a new value.
	// +optional
	UpdateRequests int `json:"updateRequests,omitempty"`

	// Spec contains the configuration of the Build to create for the App.
	// +optional
	Spec *BuildSpec `json:"spec,omitempty"`

	// Image is a ready-to-go container image to use instead of the Build.
	// +optional
	Image *string `json:"image,omitempty"`

	// BuildRef references an ADX Build (group:
	// builds.appdevexperience.dev) that should be used instead of a Build.
	// NOTE: This API is currently in preview.
	// +optional
	BuildRef *corev1.LocalObjectReference `json:"buildRef,omitempty"`
}

// AppSpecTemplate defines an app's runtime configuration.
type AppSpecTemplate struct {

	// UpdateRequests is a unique identifier for an AppSpecTemplate.
	// Updating sub-values will trigger a new value.
	UpdateRequests int `json:"updateRequests"`

	// Template is a PodSpec with additional restrictions.
	// The image name is ignored.
	// The Spec contains configuration for the App's Pod.
	// (Env, Vars, Quotas, etc)
	// +optional
	Spec corev1.PodSpec `json:"spec,omitempty"`
}

// AppSpecInstances defines the scaling rules for an App.
type AppSpecInstances struct {

	// Autoscaling defines an App's autoscaling configurations.
	Autoscaling AppSpecAutoscaling `json:"autoscaling,omitempty"`

	// Stopped determines if the App should be running or not.
	Stopped bool `json:"stopped,omitempty"`

	// Replicas defines a static number of desired instances.
	Replicas *int32 `json:"replicas,omitempty"`

	// DeprecatedExactly value is copied to Replicas.
	DeprecatedExactly *int32 `json:"exactly,omitempty"`
}

// AppSpecAutoscaling defines the autoscaling specs for an App.
type AppSpecAutoscaling struct {

	// Enabled determines if the App should have autoscaling enabled or not.
	Enabled bool `json:"enabled,omitempty"`

	// MinReplicas defines the minimum number of desired instances.
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// MaxReplicas defines the maximum number of desired instances.
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`

	// Rules defines the autoscaling rules for the App.
	Rules []AppAutoscalingRule `json:"rules,omitempty"`
}

// AutoscalingRuleType defines supported ruletypes for autoscaling.
type AutoscalingRuleType string

// Allowed RuleTypes rule autoscaling.
const (
	CPURuleType AutoscalingRuleType = "CPU"
)

// AppAutoscalingRule defines the autoscaling rules for an App.
type AppAutoscalingRule struct {

	// RuleType is the name of the scaling rule (e.g., CPU).
	RuleType AutoscalingRuleType `json:"ruleType,omitempty"`

	// Target value for the metric.
	// Unit of target depends on the rule type.
	// For CPU, it will be a percentage represented by number in range (0, 100].
	Target *int32 `json:"target,omitempty"`
}

// AutoscalingRuleStatus contains the current status of an autoscaling rule.
type AutoscalingRuleStatus struct {
	AppAutoscalingRule `json:",inline"`

	// XXX: The JSON tag for the next field is capitalized because golang
	// matches field casing by default and Kf shipped with this field untagged
	// originally so modifying it would be a breaking change.

	Current AutoscalingRuleMetricValueStatus `json:"Current"`
}

// AutoscalingRuleMetricValueStatus stores the metric value status.
//
// TODO: This closely resembles
// https://godoc.org/k8s.io/api/autoscaling/v2beta2#MetricValueStatus
// Once the v2betaX is promoted to v2, it should be replaced.
type AutoscalingRuleMetricValueStatus struct {
	AverageValue *resource.Quantity `json:"averageValue,omitempty"`
}

// RequiresHPA determines if HPA needed to be created an app.
func (autoscaling *AppSpecAutoscaling) RequiresHPA() bool {
	return autoscaling.Enabled && autoscaling.MaxReplicas != nil && len(autoscaling.Rules) > 0
}

// GetAutoscalingRuleType converts a string to AutoscalingRuleType.
// This is used by CLI to get rule type based on string.
// No validation is needed.
func GetAutoscalingRuleType(s string) AutoscalingRuleType {
	return AutoscalingRuleType(strings.ToUpper(s))
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Scale psedo-implements the autoscaling/v1 Scale type. It wraps the the type
// with Validators and Defaulters so that the webhooks can utilize it.
type Scale struct {
	autoscaling.Scale
}

var _ apis.Validatable = (*Scale)(nil)
var _ apis.Defaultable = (*Scale)(nil)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScaleList is a list of Scale resources.
type ScaleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Scale `json:"items"`
}

// GetObjectKind implements runtime.Object but for autoscaling/v1.Scale.
func (s *Scale) GetObjectKind() schema.ObjectKind {
	return s.Scale.GetObjectKind()
}

// NeedsUpdateRequestsIncrement returns true if UpdateRequests needs to be
// incremented to force a rebuild. This function should be used as a part of
// defaulting and validating webhooks when BuildSpec is embedded.
//
// This can happen if a field in the build changes without also updating the
// UpdateRequests.
func (spec *AppSpecBuild) NeedsUpdateRequestsIncrement(old AppSpecBuild) bool {
	updateRequestsChanged := old.UpdateRequests != spec.UpdateRequests

	if !updateRequestsChanged {
		specsChanged := !reflect.DeepEqual(&old, spec)
		return specsChanged
	}

	return false
}

// NeedsUpdateRequestsIncrement returns true if UpdateRequests needs to be
// incremented to force a redeploy. This function should be used as a part of
// defaulting and validating webhooks when AppSpecTemplate is embedded.
//
// This can happen if a field in the spec changes without also updating the
// UpdateRequests.
func (spec *AppSpec) NeedsUpdateRequestsIncrement(old AppSpec) bool {
	updateRequestsChanged := old.Template.UpdateRequests != spec.Template.UpdateRequests

	if !updateRequestsChanged {
		specsChanged := !reflect.DeepEqual(&old, spec)
		return specsChanged
	}

	return false
}

// DeploymentReplicas returns the value that deployment replicas should be set
// to.
func (instances *AppSpecInstances) DeploymentReplicas() (int32, error) {
	switch {
	case instances.Stopped:
		return 0, nil
	case instances.Replicas != nil:
		return *instances.Replicas, nil
	default:
		return 0, errors.New("Exact scale required for deployment based setup")
	}
}

// Status returns an InstanceStatus representing this AppSpecInstancces
// indicating what the real scale factor was.
func (instances *AppSpecInstances) Status() InstanceStatus {
	out := InstanceStatus{}

	switch {
	case instances.Stopped:
		out.Representation = "stopped"
		out.Replicas = 0
	case instances.Replicas != nil:
		exactly := fmt.Sprintf("%d", *instances.Replicas)
		out.Representation = exactly
		out.Replicas = *instances.Replicas
	default:
		out.Representation = "1"
		out.Replicas = 1
	}

	if !instances.Stopped && instances.Autoscaling.RequiresHPA() {
		autoscalingStatus := fmt.Sprintf(" (autoscaled %d to %d)", *instances.Autoscaling.MinReplicas, *instances.Autoscaling.MaxReplicas)
		out.Representation += autoscalingStatus
	}

	return out
}

// AppStatus is the current configuration and running state for an App.
type AppStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	// BuildStatusFields embeds the image and build name for the latest
	// passing build.
	BuildStatusFields `json:",inline"`

	// LatestReadyBuildName contains the name of the build that was most
	// recently built correctly.
	LatestReadyBuildName string `json:"latestReadyBuild,omitempty"`

	// LatestCreatedBuildName contains the name of the build that was most
	// recently created.
	LatestCreatedBuildName string `json:"latestBuild,omitempty"`

	// ServiceBindings are the bindings currently attached to the App.
	ServiceBindingNames []string `json:"serviceBindings,omitempty"`

	// ServiceBindingConditions are the conditions of the service bindings.
	ServiceBindingConditions duckv1beta1.Conditions `json:"serviceBindingConditions"`

	// Routes contains the statuses of the Routes attached to the instance.
	Routes []AppRouteStatus `json:"routes,omitempty"`

	// URLs is an aggregated list of URLs from Routes. This is a vanity field
	// that is necessary due to:
	// https://github.com/kubernetes/kubernetes/issues/67268
	//
	// The values found here are also found in .status.routes[*].url
	//
	// NOTE: All Route URLs will be included. This includes non-ready ones.
	URLs []string `json:"urls,omitempty"`

	// RouteConditions are the conditions of the routes.
	RouteConditions duckv1beta1.Conditions `json:"routeConditions,omitempty"`

	// Volumes are the conditions of the volumes.
	Volumes []AppVolumeStatus `json:"volumes,omitempty"`

	// Instances are actual status of the instance counts.
	Instances InstanceStatus `json:"instances,omitempty"`

	// ServiceAccountName is the service account used by the app.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Tasks are status of Tasks run on the App.
	Tasks AppTaskStatus `json:"tasks,omitempty"`

	// StartCommands are the container and buildpack start commands.
	StartCommands StartCommandStatus `json:"startCommands,omitempty"`
}

// StartCommandStatus contains the app start commands.
type StartCommandStatus struct {
	Container []string `json:"container,omitempty"`
	Buildpack []string `json:"buildpack,omitempty"`
	Image     string   `json:"image,omitempty"`
	Error     string   `json:"error,omitempty"`
}

// AppVolumeStatus contains the status of mounted volume.
type AppVolumeStatus struct {
	MountPath       string `json:"mountPath"`
	VolumeName      string `json:"name"`
	VolumeClaimName string `json:"claim"`
	ReadOnly        bool   `json:"readonly,omitempty"`

	UidGid `json:",inline"`
}

// InstanceStatus contains the computed scaling.
type InstanceStatus struct {
	// Replicas contains the number of App instances.
	Replicas int32 `json:"replicas,omitempty"`

	// Representation contains a human readable description of the instance
	// status.
	Representation string `json:"representation,omitempty"`

	// LabelSelector for pods. It must match the pod template's labels.
	LabelSelector string `json:"labelSelector"`

	// AutoscalingStatus contains status for each autoscaling rule
	AutoscalingStatus []AutoscalingRuleStatus `json:"autoscalingStatus,omitempty"`

	// DeprecatedEffectiveMin contains the effective minimum number of
	// instances passed as an annotation value.
	DeprecatedEffectiveMin string `json:"effectiveMin,omitempty"`

	// DeprecatedEffectiveMax contains the effective maximum number of
	// instances passed as an annotation.
	DeprecatedEffectiveMax string `json:"effectiveMax,omitempty"`
}

// AppTaskStatus contains the Task status on the App.
type AppTaskStatus struct {
	// UpdateRequests contains the number of run-task requests on the App.
	UpdateRequests int `json:"updateRequests"`
}

// AppRouteStatus contains the status information about a Route.
type AppRouteStatus struct {
	QualifiedRouteBinding `json:",inline"`

	// VirtualService is the VirtualService that is created with the Route.
	VirtualService corev1.LocalObjectReference `json:"virtualservice,omitempty"`

	// URL is the URL for the route
	URL string `json:"url,omitempty"`

	// Status contains the status of this binding.
	Status RouteBindingStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppList is a list of App resources.
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []App `json:"items"`
}
