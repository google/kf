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
	"context"
	"strings"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/networking"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/reconciler/build/config"
	"github.com/google/kf/v2/pkg/system"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/apis"
)

var (
	// appStagingMaxScratchSpace contains the upper bound bound for the amount of
	// disk used by CF's buildpack staging process:
	// https://docs.cloudfoundry.org/devguide/deploy-apps/troubleshoot-app-health.html#upload
	appStagingMaxScratchSpace = resource.MustParse("4Gi")
)

// ClusterDiagnostic tests that the cluster's Kubernetes version and components
// are suitable for kf.
type ClusterDiagnostic struct {
	kubeClient kubernetes.Interface
}

type expectedDeployments struct {
	Namespace   string
	Deployments []string
}

type expectedComponent struct {
	groupVersion      string
	expectedResources []string
}

var _ Diagnosable = (*ClusterDiagnostic)(nil)

// Diagnose validates the version and components of the current Kubernetes
// cluster.
func (c *ClusterDiagnostic) Diagnose(ctx context.Context, d *Diagnostic) {
	d.Run(ctx, "Version", func(ctx context.Context, d *Diagnostic) {
		diagnoseKubernetesVersion(ctx, d, c.kubeClient.Discovery())
	})

	d.Run(ctx, "Controllers", func(ctx context.Context, d *Diagnostic) {
		diagnoseControllers(ctx, d, c.kubeClient)
	})

	d.Run(ctx, "Components", func(ctx context.Context, d *Diagnostic) {
		diagnoseKfComponents(ctx, d, c.kubeClient.Discovery())
	})

	d.Run(ctx, "MutatingWebhooks", func(ctx context.Context, d *Diagnostic) {
		diagnoseMutatingWebhooks(ctx, d, c.kubeClient)
	})

	d.Run(ctx, "Ingresses", func(ctx context.Context, d *Diagnostic) {
		diagnoseIngress(ctx, d, c.kubeClient)
	})

	d.Run(ctx, "DaemonSets", func(ctx context.Context, d *Diagnostic) {
		diagnoseDaemonSets(ctx, d, c.kubeClient)
	})

	d.Run(ctx, "Nodes", func(ctx context.Context, d *Diagnostic) {
		diagnoseNodes(ctx, d, c.kubeClient)
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
func diagnoseKubernetesVersion(ctx context.Context, d *Diagnostic, vc discovery.ServerVersionInterface) {
	k8sVersion, err := vc.ServerVersion()

	testutil.AssertNil(d, "server version error", err)
	testutil.AssertEqual(d, "major version", "1", k8sVersion.Major)
	testutil.AssertEqual(d, "platform", "linux/amd64", k8sVersion.Platform)
}

// diagnoseComponents checks that the Kubernetes instance has all
// the required components to run kf.
func diagnoseKfComponents(ctx context.Context, d *Diagnostic, vc discovery.ServerResourcesInterface) {
	expectedComponents := map[string]expectedComponent{
		"Kubernetes V1": {
			groupVersion:      "v1",
			expectedResources: []string{"configmaps", "secrets", "resourcequotas"},
		},
		"Tekton": {
			groupVersion:      "tekton.dev/v1beta1",
			expectedResources: []string{"taskruns", "tasks", "clustertasks"},
		},
	}

	diagnoseComponents(expectedComponents, ctx, d, vc)
}

// diagnoseComponents checks that the Kubernetes instance has all
// the required components to run kf.
func diagnoseComponents(expectedComponents map[string]expectedComponent, ctx context.Context, d *Diagnostic, vc discovery.ServerResourcesInterface) {
	for tn, tc := range expectedComponents {
		d.Run(ctx, tn, func(ctx context.Context, d *Diagnostic) {
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
				d.Run(ctx, resource, func(ctx context.Context, d *Diagnostic) {
					if !foundResources[resource] {
						d.Errorf("Expected to find resource %s", resource)
					}
				})
			}
		})
	}
}

func diagnoseDeployments(deployments []expectedDeployments, ctx context.Context, d *Diagnostic, kubernetes kubernetes.Interface) {
	for _, tc := range deployments {
		d.Run(ctx, tc.Namespace, func(ctx context.Context, d *Diagnostic) {
			names := sets.NewString()

			deployments, err := kubernetes.AppsV1().Deployments(tc.Namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				d.Errorf("couldn't list deployments: %v", err)
				return
			}

			if len(deployments.Items) == 0 {
				d.Fatalf("namespace %v has no deployments", tc.Namespace)
				return
			}

			for _, deployment := range deployments.Items {
				names.Insert(deployment.Name)

				d.Run(ctx, deployment.Name, func(ctx context.Context, d *Diagnostic) {
					ready := deployment.Status.ReadyReplicas
					desired := deployment.Status.Replicas

					if ready != desired {
						d.Errorf("ready: %d/%d", ready, desired)
					}
				})
			}

			diff := sets.NewString(tc.Deployments...).Difference(names)
			if diff.Len() != 0 {
				d.Errorf("missing expected deployment(s): %v have: %v", diff.List(), names.List())
			}
		})
	}
}

func diagnoseControllers(ctx context.Context, d *Diagnostic, kubernetes kubernetes.Interface) {
	isADXBuildsEnabled := adxBuildsEnabled(ctx, d, kubernetes)

	deployments := []expectedDeployments{
		{
			Namespace:   "kf",
			Deployments: []string{"webhook", "controller", "subresource-apiserver"},
		},
		{
			Namespace:   "tekton-pipelines",
			Deployments: []string{"tekton-pipelines-controller", "tekton-pipelines-webhook"},
		},
	}

	if isADXBuildsEnabled {
		deployments = append(deployments,
			expectedDeployments{
				Namespace:   "adx-builds-system",
				Deployments: []string{"webhook", "controller", "subresource-apiserver"},
			})
	}

	// We only require KCC if WI is enabled AND ADX Builds are not being used.
	{
		if workloadIdentityEnabled(ctx, d, kubernetes) && !isADXBuildsEnabled {
			// WI field populated, assume WI is enabled.
			deployments = append(deployments, expectedDeployments{
				Namespace:   "cnrm-system",
				Deployments: []string{"cnrm-resource-stats-recorder", "cnrm-webhook-manager"},
			})
		} else {
			// Doesn't look like WI is used or ADX builds is enabled, move on.
			d.Log("WI is not enabled or AppDeveloperExperience Builds is enabled, skipping KCC/CNRM checks")
		}
	}

	// We only expect an Istio deployment if not using MCP.
	{
		isManaged, err := networking.IsManagedASM(func(namespace, configMapName string) (*corev1.ConfigMap, error) {
			return kubernetes.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		})
		if err != nil {
			d.Warnf("failed to detect ASM install configuration: %v", err)
		}
		if !isManaged {
			deployments = append(deployments, expectedDeployments{
				Namespace: "istio-system",
				// Don't use specific deployment names as they may differ between versions.
				Deployments: []string{},
			})
		}
	}

	diagnoseDeployments(deployments, ctx, d, kubernetes)
}

func workloadIdentityEnabled(
	ctx context.Context,
	d *Diagnostic,
	kubernetes kubernetes.Interface,
) bool {
	configSecrets, err := kubernetes.CoreV1().ConfigMaps("kf").Get(
		ctx,
		config.SecretsConfigName,
		metav1.GetOptions{},
	)
	if err != nil {
		d.Fatalf("failed to get Kf secret: %v", err)
	}

	_, ok := configSecrets.Data[config.GoogleServiceAccountKey]
	return ok
}

func adxBuildsEnabled(
	ctx context.Context,
	d *Diagnostic,
	kubernetes kubernetes.Interface,
) bool {
	cm, err := kubernetes.CoreV1().ConfigMaps("kf").Get(
		ctx,
		kfconfig.DefaultsConfigName,
		metav1.GetOptions{},
	)
	if err != nil {
		d.Fatalf("failed to get Kf defaults: %v", err)
	}

	configDefaults, err := kfconfig.NewDefaultsConfigFromConfigMap(cm)
	if err != nil {
		d.Fatalf("failed to parse config-defaults: %v", err)
	}

	return configDefaults.FeatureFlags.AppDevExperienceBuilds().IsEnabled()
}

func diagnoseMutatingWebhooks(ctx context.Context, d *Diagnostic, kubernetes kubernetes.Interface) {
	expected := []struct {
		Component string
		Webhooks  []string
	}{
		{
			Component: "kf",
			Webhooks:  []string{"webhook.kf.dev"},
		},
	}

	for _, tc := range expected {
		d.Run(ctx, tc.Component, func(ctx context.Context, d *Diagnostic) {

			for _, name := range tc.Webhooks {
				d.Run(ctx, name, func(ctx context.Context, d *Diagnostic) {

					_, err := kubernetes.
						AdmissionregistrationV1().
						MutatingWebhookConfigurations().
						Get(ctx, name, metav1.GetOptions{})
					if err != nil {
						d.Errorf("couldn't fetch mutatingWebhook %s: %v", name, err)
						return
					}
				})
			}
		})
	}
}

type serviceListerAdapter struct {
	kubernetes kubernetes.Interface
}

// List implements system.ServiceLister
func (s *serviceListerAdapter) List(selector labels.Selector) ([]*corev1.Service, error) {
	serviceList, err := s.kubernetes.CoreV1().Services("").List(context.Background(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}

	out := []*corev1.Service{}
	for i := range serviceList.Items {
		out = append(out, &serviceList.Items[i])
	}

	return out, nil
}

var _ system.ServiceLister = (*serviceListerAdapter)(nil)

func diagnoseIngress(ctx context.Context, d *Diagnostic, kubernetes kubernetes.Interface) {
	isManaged, err := networking.IsManagedASM(func(namespace, configMapName string) (*corev1.ConfigMap, error) {
		return kubernetes.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
	})
	if err != nil {
		d.Warnf("couldn't detect ASM install configuration: %v", err)
	}

	var helpSuffix string
	if err == nil && isManaged {
		helpSuffix = `Ensure you have created an Istio ingress gateway and labeled
the service to match the selector. Check the Google-managed
control plane ASM install documentation for an example:

    https://cloud.google.com/service-mesh/docs/managed-control-plane#install_istio_gateways_optional
`
	} else {
		helpSuffix = `If this step repeatedly fails, check your GCP quota for:
* Load balancers
* IP addresses
* Firewall rules

    https://console.cloud.google.com/iam-admin/quotas
`
	}
	// The following only gets shown to the user if the tests fail.
	d.Logf(`Looking up ingress services, these should match the selector in the Kf ingress gateway
Use the following command to see the existing selector:
    kubectl -n kf get gateways external-gateway -o yaml

%s`, helpSuffix)

	adapter := &serviceListerAdapter{
		kubernetes: kubernetes,
	}

	ingresses, err := system.GetClusterIngresses(adapter)
	switch {
	case err != nil:
		d.Fatalf("couldn't look up ingresses: %v", err)
	case len(ingresses) == 0:
		d.Fatal("no ingresses could be found")
	default:
		if _, err := system.ExtractProxyIngressFromList(ingresses); err != nil {
			d.Fatalf("no valid ingresses were found: %v", err)
		}
	}
}

// Validates DaemonSets that are in controller namespaces, with a PASS result if there are no DaemonSets, or if
// the desired scheduled daemon pods, current scheduled daemon pods, and currently running daemon pods are the same.
// See https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#daemonsetstatus-v1-apps for DaemonSet
// documentation, and https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/kubectl/pkg/polymorphichelpers/rollout_status.go#L95
// for how kubectl handles the DaemonSet status when running `kubectl get daemonsets`
func diagnoseDaemonSets(ctx context.Context, d *Diagnostic, kubernetes kubernetes.Interface) {
	namespaces := sets.NewString("kf", "kube-system")
	namespaces.Insert(DiscoverControllerNamespaces(ctx, kubernetes)...)
	for ns := range namespaces {
		d.Run(ctx, ns, func(ctx context.Context, d *Diagnostic) {
			daemonsets, err := kubernetes.AppsV1().DaemonSets(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				d.Errorf("Could not retrieve DaemonSets for namespace %s: %v\n", ns, err)
			} else if daemonsets == nil || len(daemonsets.Items) == 0 {
				d.Logf("Did not find DaemonSets for namespace %s.\n", ns)
				return
			}
			for _, tc := range daemonsets.Items {
				d.Run(ctx, tc.Name, func(ctx context.Context, d *Diagnostic) {
					formatString := "DaemonSet %s has %d out of %d desired nodes with scheduled daemon pods, and %d " +
						"nodes have the daemon pod running and ready.\n"
					desired := tc.Status.DesiredNumberScheduled
					current := tc.Status.CurrentNumberScheduled
					ready := tc.Status.NumberReady

					if current < desired || ready < desired {
						d.Fatalf(formatString, tc.Name, current, desired, ready)
					}
				})
			}
		})
	}
}

func diagnoseNodes(ctx context.Context, d *Diagnostic, kubernetes kubernetes.Interface) {
	nodesList, err := kubernetes.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	switch {
	case err != nil:
		d.Warnf("couldn't look up ingresses: %v", err)
	case len(nodesList.Items) == 0:
		d.Warn("no Nodes were found")
	default:
		nodes := nodesList.Items
		if nodeCount := len(nodes); nodeCount < 4 {
			d.Warnf("Kf recommends clusters with at least 4 nodes, found %d", nodeCount)
		}

		for _, node := range nodes {
			d.Run(ctx, node.Name, func(ctx context.Context, d *Diagnostic) {

				// Check that Calico exists for NetworkPolicies; the non-existence of
				// Calico is an indicator that things are wrong, but isn't a guarantee.
				d.Run(ctx, "calico-enabled", func(ctx context.Context, d *Diagnostic) {
					val, ok := node.Labels["projectcalico.org/ds-ready"]
					if !ok || val != "true" {
						d.Warn("Calico appears not to be enabled; NetworkPolicies may not work for workloads on this node.")
					}
				})

				if workloadIdentityEnabled(ctx, d, kubernetes) {
					// Check that the metadataserver label exists. If it doesn't the Node
					// won't work with WorkloadIdentity.
					d.Run(ctx, "gke-metadata-server-enabled", func(ctx context.Context, d *Diagnostic) {
						val, ok := node.Labels["iam.gke.io/gke-metadata-server-enabled"]
						if !ok || val != "true" && workloadIdentityEnabled(ctx, d, kubernetes) {
							d.Fatal("The GKE Metadata Server (Workload Identity) appears not to be enabled on this node.")
						}
					})
				}

				d.Run(ctx, "properties", func(ctx context.Context, d *Diagnostic) {
					nodeInfo := node.Status.NodeInfo

					// NOTE: The values here appear to follow goarch values for GCE nodes
					// but the API spec doesn't explain the full range of values they
					// can report. That and the fact that the K8s scheduler should not
					// schedule Pods onto incompatible Kubelets make these warnings rather
					// than failures.

					if !strings.EqualFold("amd64", nodeInfo.Architecture) {
						d.Warnf("Kf only supports amd64 architectures, got %q", nodeInfo.Architecture)
					}

					if !strings.EqualFold("linux", nodeInfo.OperatingSystem) {
						d.Warnf("Kf only supports Linux operating systems, got %q", nodeInfo.OperatingSystem)
					}
				})

				d.Run(ctx, "ephemeral storage", func(ctx context.Context, d *Diagnostic) {
					if ephemeralStorage, ok := node.Status.Allocatable[corev1.ResourceEphemeralStorage]; ok {
						if ephemeralStorage.Cmp(appStagingMaxScratchSpace) < 0 {
							d.Warnf(
								"Kf needs at least %s of ephemeral storage to build Apps, node only has %s",
								appStagingMaxScratchSpace.String(),
								ephemeralStorage.String(),
							)
						}
					}
				})

				d.Run(ctx, "conditions", func(ctx context.Context, d *Diagnostic) {
					for _, cond := range node.Status.Conditions {
						switch {
						// The Ready condition should be true indicating the node is healthy.
						case apis.ConditionType(cond.Type) == apis.ConditionReady:
							if cond.Status != corev1.ConditionTrue {
								d.Errorf(
									"Node is not ready, current status is %q with message %q",
									cond.Status,
									cond.Message,
								)
							}
						default:
							// Other conditions on the node should be false because they
							// indicate specific failure conditions when set to true.
							if cond.Status != corev1.ConditionFalse {
								d.Errorf(
									"Condition %s is not healthy, current status is %q with message %q",
									cond.Type,
									cond.Status,
									cond.Message,
								)
							}
						}
					}
				})
			})
		}
	}
}
