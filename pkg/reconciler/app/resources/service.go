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

package resources

import (
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfv1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/kmeta"
)

const (
	// UserPortName is the arbitrary name given to the port the container will
	// listen on.
	UserPortName = "http-user-port"

	// DefaultUserPort is the default port for a container to listen on.
	DefaultUserPort = 8080
)

// PodLabels returns the labels for selecting pods of the deployment.
func PodLabels(app *v1alpha1.App) map[string]string {
	return app.ComponentLabels("app-server")
}

// ServiceName is the name of the service for the app
func ServiceName(app *kfv1alpha1.App) string {
	return ServiceNameForAppName(app.Name)
}

// ServiceNameForAppName returns the canonical service name to route requests
// to the given app.
func ServiceNameForAppName(appName string) string {
	return v1alpha1.GenerateName(appName)
}

// MakeService constructs a K8s service, that is backed by the pod selector
// matching pods created by the revision.
func MakeService(app *kfv1alpha1.App) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName(app),
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: v1alpha1.UnionMaps(app.GetLabels(), app.ComponentLabels("service")),
		},
		Spec: corev1.ServiceSpec{
			Ports:    makeServicePorts(app),
			Selector: PodLabels(app),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

// makeServicePorts creates ports on the Kubernetes Service for each exposed port
// on the App container. If no user-defined port has the name http-user-port or
// the port number 80, a default ServicePort is injected. This value allows
// existing Istio VirtualServices to continue working.
func makeServicePorts(app *kfv1alpha1.App) (ports []corev1.ServicePort) {
	containers := app.Spec.Template.Spec.Containers
	injectDefaultPort := true
	if len(containers) != 0 {
		for _, containerPort := range containers[0].Ports {
			// don't inject the default port if there's a conflicting port
			if containerPort.Name == UserPortName || containerPort.ContainerPort == 80 {
				injectDefaultPort = false
			}
			ports = append(ports, corev1.ServicePort{
				Name:       containerPort.Name,
				Protocol:   containerPort.Protocol,
				Port:       containerPort.ContainerPort,
				TargetPort: intstr.FromInt(int(containerPort.ContainerPort)),
			})
		}
	}

	// TODO(jlewisiii): in v1beta1 this injection shoud be removed because it's
	// for reverse compatibility.
	if injectDefaultPort {
		ports = append(ports, corev1.ServicePort{
			Name:     UserPortName,
			Protocol: corev1.ProtocolTCP,
			Port:     v1alpha1.DefaultRouteDestinationPort,
			// This one is matching the public one, since this is the
			// port queue-proxy listens on.
			TargetPort: intstr.FromInt(int(getUserPort(app))),
		})
	}

	return
}

func getUserPort(app *v1alpha1.App) int32 {
	containers := app.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return DefaultUserPort
	}

	ports := containers[0].Ports

	if len(ports) > 0 && ports[0].ContainerPort != 0 {
		return ports[0].ContainerPort
	}

	// TODO: Consider using container EXPOSE metadata from image before
	// falling back to default value.

	return DefaultUserPort
}
