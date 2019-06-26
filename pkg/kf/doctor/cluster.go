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

package doctor

import (
	"github.com/google/kf/pkg/kf/testutil"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
)

// ClusterDiagnostic tests that the cluster's Kubernetes version and components
// are suitable for kf.
type ClusterDiagnostic struct {
	kubeClient kubernetes.Interface
}

var _ Diagnosable = (*ClusterDiagnostic)(nil)

// Diagnose valiates the version and components of the current Kubernetes
// cluster.
func (c *ClusterDiagnostic) Diagnose(d *Diagnostic) {
	d.Run("Version", func(d *Diagnostic) {
		diagnoseKubernetesVersion(d, c.kubeClient.Discovery())
	})

	d.Run("Components", func(d *Diagnostic) {
		diagnoseComponents(d, c.kubeClient.Discovery())
	})
}

// NewClusterDiagnostic creates a new ClusterDiagnostic to validate the
// install pointed at by the client.
func NewClusterDiagnostic(kubeClient kubernetes.Interface) *ClusterDiagnostic {
	return &ClusterDiagnostic{
		kubeClient: kubeClient,
	}
}

// diagnoseKubernetesVersion checks the version and platform k8s is running on
// is compatible with kf.
func diagnoseKubernetesVersion(d *Diagnostic, vc discovery.ServerVersionInterface) {
	k8sVersion, err := vc.ServerVersion()

	testutil.AssertNil(d, "server version error", err)
	testutil.AssertEqual(d, "major version", "1", k8sVersion.Major)
	testutil.AssertEqual(d, "platform", "linux/amd64", k8sVersion.Platform)
}

// diagnoseComponents checks that the Kubernetes instance has all
// the required components to run kf.
func diagnoseComponents(d *Diagnostic, vc discovery.ServerResourcesInterface) {
	expectedComponents := map[string]struct {
		groupVersion      string
		expectedResources []string
	}{
		"Knative Serving": {
			groupVersion:      "serving.knative.dev/v1alpha1",
			expectedResources: []string{"configurations", "routes", "revisions", "services"},
		},
		"Service Catalog": {
			groupVersion:      "servicecatalog.k8s.io/v1beta1",
			expectedResources: []string{"clusterservicebrokers"},
		},
		"Kubernetes V1": {
			groupVersion:      "v1",
			expectedResources: []string{"configmaps", "secrets", "resourcequotas"},
		},
	}

	for tn, tc := range expectedComponents {
		d.Run(tn, func(d *Diagnostic) {
			resourceList, err := vc.ServerResourcesForGroupVersion(tc.groupVersion)
			if err != nil {
				d.Fatalf("Error getting resources for %s: %v", tc.groupVersion, err)
			}

			foundResources := make(map[string]bool)
			for _, r := range resourceList.APIResources {
				foundResources[r.Name] = true
				d.Log("found resource", r.Name)
			}

			for _, resource := range tc.expectedResources {
				d.Run(resource, func(d *Diagnostic) {
					if !foundResources[resource] {
						d.Errorf("Expected to find resource %s", resource)
					}
				})
			}
		})
	}
}
