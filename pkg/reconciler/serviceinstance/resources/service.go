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
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/kmeta"
)

// ServiceName is the name of the K8s Service for the route service proxy
func ServiceName(serviceInstance *v1alpha1.ServiceInstance) string {
	return ServiceNameForRouteServiceName(serviceInstance.Name)
}

// ServiceNameForRouteServiceName returns the canonical service name for the Service that
// directs requests to the route service.
func ServiceNameForRouteServiceName(serviceInstanceName string) string {
	return v1alpha1.GenerateName(serviceInstanceName, "proxy")
}

// MakeService constructs a K8s service for the route service proxy.
// The Service is backed by the pod selector matching pods created by the deployment.
func MakeService(serviceInstance *v1alpha1.ServiceInstance) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName(serviceInstance),
			Namespace: serviceInstance.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(serviceInstance),
			},
			Labels: v1alpha1.UnionMaps(serviceInstance.GetLabels()),
		},
		Spec: corev1.ServiceSpec{
			Ports:    makeServicePorts(serviceInstance),
			Selector: PodLabels(serviceInstance),
		},
	}
}

// makeServicePorts creates a port on the K8s Service that exposes port 80 (for HTTP traffic)
// and targets the port defined in the PodSpec of the route service proxy deployment.
func makeServicePorts(serviceInstance *v1alpha1.ServiceInstance) []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:       RouteServiceProxyUserPortName,
			Protocol:   corev1.ProtocolTCP,
			Port:       v1alpha1.DefaultRouteDestinationPort,
			TargetPort: intstr.FromInt(int(RouteServiceProxyUserPort)),
		},
	}
}
