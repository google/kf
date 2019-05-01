package doctor

import (
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
)

// KubernetesClientFactory creates a Kubernetes client
type KubernetesClientFactory func() (kubernetes.Interface, error)

// ClusterDiagnostic tests that the cluster's Kubernetes version and components
// are suitable for kf.
type ClusterDiagnostic struct {
	createKubernetesClient KubernetesClientFactory
}

var _ Diagnosable = (*ClusterDiagnostic)(nil)

// Diagnose valiates the version and components of the current Kubernetes
// cluster.
func (c *ClusterDiagnostic) Diagnose(d *Diagnostic) {
	client, err := c.createKubernetesClient()
	testutil.AssertNil(d, "Kubernetes Client Error", err)

	d.Run("Version", func(d *Diagnostic) {
		diagnoseKubernetesVersion(d, client.Discovery())
	})

	d.Run("Components", func(d *Diagnostic) {
		diagnoseComponents(d, client.Discovery())
	})
}

// NewClusterDiagnostic creates a new ClusterDiagnostic to validate the
// install pointed at by the client.
func NewClusterDiagnostic(clientFactory KubernetesClientFactory) *ClusterDiagnostic {
	return &ClusterDiagnostic{
		createKubernetesClient: clientFactory,
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
			expectedResources: []string{"configmaps", "secrets"},
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
