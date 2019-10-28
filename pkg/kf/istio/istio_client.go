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

package istio

import (
	"errors"

	"github.com/google/kf/pkg/kf/doctor"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go istio_client_options.yml istio_client_options.go

// IngressLister gets Istio ingresses points for clusters.
type IngressLister interface {
	doctor.Diagnosable

	ListIngresses(opts ...ListIngressesOption) ([]corev1.LoadBalancerIngress, error)
}

// NewIstioClient creates an IngressLister used to find ingress points for the
// cluster's apps.
func NewIstioClient(c kubernetes.Interface) IngressLister {
	return &istioClient{c: c}
}

// istioClient is used to find ingress points for the cluster's apps.
type istioClient struct {
	c kubernetes.Interface
}

// ListIngresses gets the Istio ingres IP address(es) for a particular gateway.
func (c *istioClient) ListIngresses(opts ...ListIngressesOption) ([]corev1.LoadBalancerIngress, error) {
	cfg := ListIngressesOptionDefaults().Extend(opts).toConfig()

	svc, err := c.c.
		CoreV1().
		Services(cfg.Namespace).
		Get(cfg.Service, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return svc.Status.LoadBalancer.Ingress, nil
}

// Diagnose implements doctor.Diagnosable
func (c *istioClient) Diagnose(d *doctor.Diagnostic) {
	d.Run("Istio Ingress", func(d *doctor.Diagnostic) {
		ingresses, err := c.ListIngresses()
		if err != nil {
			d.Fatalf("couldn't find cluster ingress: %s", err)
			return
		}

		if len(ingresses) == 0 {
			d.Errorf("no public ingresses found")
		}
	})
}

// ExtractIngressFromList is a utility function to wrap IngressLister.ListIngresses
// and extract a single ingress.
func ExtractIngressFromList(ingresses []corev1.LoadBalancerIngress, err error) (string, error) {
	if err != nil {
		return "", err
	}

	if len(ingresses) == 0 {
		return "", errors.New("no ingresses were found")
	}

	for _, ingress := range ingresses {
		if ingress.IP != "" {
			return ingress.IP, nil
		}
	}

	return "", errors.New("no ingresses had IP addresses listed")
}
