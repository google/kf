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
	"fmt"

	"github.com/knative/serving/pkg/apis/autoscaling"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"

	core "k8s.io/api/core/v1"
)

const (
	// NameLabel holds the standard label key for Kubernetes app names.
	NameLabel = "app.kubernetes.io/name"
	// ManagedByLabel holds the standard label key for Kubernetes app managers.
	ManagedByLabel = "app.kubernetes.io/managed-by"
	// ComponentLabel holds the standard label key for Kubernetes app component
	// identifiers.
	ComponentLabel = "app.kubernetes.io/component"
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

// ComponentLabels returns Kubernetes recommended labels to tie together
// deployed applications and their pieces. The provided component name
// specifies the sub-resource of the app e.g. "database", "load-balancer",
// or "server".
func (app *App) ComponentLabels(component string) map[string]string {
	return map[string]string{
		NameLabel:      app.Name,
		ManagedByLabel: "kf",
		ComponentLabel: component,
	}
}

// AppSpec is the desired configuration for an App.
type AppSpec struct {

	// Source contains the source configuration of the App.
	// +optional
	Source SourceSpec `json:"source,omitempty"`

	// Template defines the App's runtime configuration.
	// +optional
	Template AppSpecTemplate `json:"template"`

	// Instances defines the scaling rules for the App.
	Instances AppSpecInstances `json:"instances,omitempty"`
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
	Spec core.PodSpec `json:"spec,omitempty"`
}

// AppSpecInstances defines the scaling rules for an App.
type AppSpecInstances struct {

	// Stopped determines if the App should be running or not.
	Stopped bool `json:"stopped,omitempty"`

	// Exactly defines a static number of desired instances.
	// If Exactly is set, it supersedes the Min and Max values.
	Exactly *int `json:"exactly,omitempty"`

	// Min defines a minimum auto-scaling limit.
	Min *int `json:"min,omitempty"`

	// Max defines a maximum auto-scaling limit.
	Max *int `json:"max,omitempty"`
}

// MinAnnotationValue returns the value autoscaling.knative.dev/minScale should
// be set to.
func (instances *AppSpecInstances) MinAnnotationValue() string {
	switch {
	case instances.Stopped:
		return "0"
	case instances.Exactly != nil:
		return fmt.Sprintf("%d", *instances.Exactly)
	case instances.Min != nil:
		return fmt.Sprintf("%d", *instances.Min)
	default:
		return ""
	}
}

// MaxAnnotationValue returns the value autoscaling.knative.dev/maxScale should
// be set to.
func (instances *AppSpecInstances) MaxAnnotationValue() string {
	switch {
	case instances.Stopped:
		return "0"
	case instances.Exactly != nil:
		return fmt.Sprintf("%d", *instances.Exactly)
	case instances.Max != nil:
		return fmt.Sprintf("%d", *instances.Max)
	default:
		return ""
	}
}

// ScalingAnnotations returns the annotations to put on the underling Serving
// to set scaling bounds.
func (instances *AppSpecInstances) ScalingAnnotations() map[string]string {
	out := make(map[string]string)

	if minVal := instances.MinAnnotationValue(); minVal != "" {
		out[autoscaling.MinScaleAnnotationKey] = minVal
	}

	if maxVal := instances.MaxAnnotationValue(); maxVal != "" {
		out[autoscaling.MaxScaleAnnotationKey] = maxVal
	}

	return out
}

// AppStatus is the current configuration and running state for an App.
type AppStatus struct {
	// Pull in the fields from Knative's duckv1beta1 status field.
	duckv1beta1.Status `json:",inline"`

	// SourceStatusFields embeds the image and build name for the latest passing
	// source.
	SourceStatusFields `json:",inline"`

	// Inline the latest serving.Service revisions that are ready
	serving.ConfigurationStatusFields `json:",inline"`

	// Inline the latest Service route information.
	serving.RouteStatusFields `json:",inline"`

	// LatestReadySourceName contains the name of the source that was most recently
	// built correctly.
	LatestReadySourceName string `json:"latestReadySource,omitempty"`

	// LatestCreatedSourceName contains the name of the source that was most recently created.
	LatestCreatedSourceName string `json:"latestSource,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppList is a list of App resources.
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []App `json:"items"`
}
