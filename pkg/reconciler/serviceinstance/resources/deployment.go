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

package resources

import (
	"errors"
	"fmt"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/ptr"
)

const (
	RouteServiceProxyUserPort     = int32(8080)
	RouteServiceProxyUserPortName = "http" // user port name must start with the protocol (http)
)

var (
	revisionHistoryLimit = int32(1)
	replicas             = int32(1)
)

// DeploymentName gets the name of a Deployment given the route service instance.
func DeploymentName(serviceInstance *v1alpha1.ServiceInstance) string {
	return v1alpha1.GenerateName(serviceInstance.Name, "proxy")
}

// PodLabels returns the labels for selecting pods of the route service proxy deployment.
func PodLabels(serviceInstance *v1alpha1.ServiceInstance) map[string]string {
	return map[string]string{
		v1alpha1.NameLabel:      fmt.Sprintf("%s-proxy", serviceInstance.Name),
		v1alpha1.ManagedByLabel: "kf",
		v1alpha1.ComponentLabel: "route-service",
	}
}

// MakeDeployment creates a K8s Deployment for a route service proxy.
func MakeDeployment(serviceInstance *v1alpha1.ServiceInstance, cfg *config.Config) (*appsv1.Deployment, error) {
	if cfg == nil {
		return nil, errors.New("the Kf defaults configmap couldn't be found")
	}
	configDefaults, err := cfg.Defaults()
	if err != nil {
		return nil, err
	}
	if configDefaults.RouteServiceProxyImage == "" {
		return nil, errors.New("config value for RouteServiceProxyImage couldn't be found")
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeploymentName(serviceInstance),
			Namespace: serviceInstance.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(serviceInstance),
			},
			Labels: v1alpha1.UnionMaps(serviceInstance.GetLabels()),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: metav1.SetAsLabelSelector(labels.Set(PodLabels(serviceInstance))),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: v1alpha1.UnionMaps(
						PodLabels(serviceInstance),

						// Insert a label for isolating apps with their own NetworkPolicies.
						map[string]string{
							v1alpha1.NetworkPolicyLabel: v1alpha1.NetworkPolicyApp,
						}),
					Annotations: map[string]string{
						"sidecar.istio.io/inject":                          "true",
						"traffic.sidecar.istio.io/includeOutboundIPRanges": "*",
					},
				},
				Spec: makePodSpec(*serviceInstance, configDefaults),
			},
			RevisionHistoryLimit: ptr.Int32(revisionHistoryLimit),
			Replicas:             ptr.Int32(replicas),
		},
	}, nil
}

func makePodSpec(serviceInstance v1alpha1.ServiceInstance, configDefaults *config.DefaultsConfig) corev1.PodSpec {
	spec := corev1.PodSpec{}
	// Don't inject the old Docker style environment variables for every service
	// doing so could cause misconfiguration for apps.
	spec.EnableServiceLinks = ptr.Bool(false)

	userContainer := corev1.Container{
		Name:            v1alpha1.DefaultUserContainerName,
		Image:           configDefaults.RouteServiceProxyImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{{
			Name:          RouteServiceProxyUserPortName,
			ContainerPort: RouteServiceProxyUserPort,
		}},
	}

	// Set environment variables on the container for Route Service URL and Port.
	routeServiceURLEnvVar := corev1.EnvVar{
		Name:  "ROUTE_SERVICE_URL",
		Value: serviceInstance.Spec.UPS.RouteServiceURL.String(),
	}
	portEnvVar := corev1.EnvVar{
		Name:  "PORT",
		Value: fmt.Sprint(RouteServiceProxyUserPort),
	}
	userContainer.Env = []corev1.EnvVar{routeServiceURLEnvVar, portEnvVar}

	// Explicitly disable stdin and tty allocation
	userContainer.Stdin = false
	userContainer.TTY = false

	// Set liveness and readiness probe TCP ports
	tcpProbe := corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(int(RouteServiceProxyUserPort)),
			},
		},
	}
	userContainer.LivenessProbe = &tcpProbe
	userContainer.ReadinessProbe = &tcpProbe
	spec.Containers = []corev1.Container{userContainer}
	return spec
}
