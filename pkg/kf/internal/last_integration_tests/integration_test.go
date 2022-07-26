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

package lastintegrationtests

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	kfconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfclient "github.com/google/kf/v2/pkg/client/kf/injection/client"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
	tektonclient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/injection/clients/dynamicclient"
)

// TestIntegration_BuildWithoutIstioSideCar updates config-defaults to have
// buildDisableIstioSidecar set to true. It then pushes an app and ensures the
// resulting build doesn't have the istio-proxy.
func TestIntegration_BuildWithoutIstioSideCar(t *testing.T) {
	integration.RunKubeAPITest(context.Background(), t, func(ctx context.Context, t *testing.T) {
		integration.RunKfTest(ctx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
			kubeClient := kubeclient.Get(ctx)
			kfSystemClient := kfSystemClient(ctx)

			// Patch Config Defaults.
			{
				integration.Logf(t, "patching config-defaults to have buildDisableIstioSidecar set to true...")
				_, err := kfSystemClient.
					Patch(
						ctx,
						"kfsystem",
						types.JSONPatchType,
						[]byte(`[{"op": "replace", "path": "/spec/kf/config/buildDisableIstioSidecar", "value": true}]`),
						metav1.PatchOptions{},
					)

				testutil.AssertErrorsEqual(t, nil, err)

				t.Cleanup(func() {
					integration.Logf(t, "setting cluster's buildDisableIstioSidecar back to false...")
					_, err := kubeClient.
						CoreV1().
						ConfigMaps("kf").
						Patch(
							// It's likely the ctx is done.
							context.Background(),
							kfconfig.DefaultsConfigName,
							types.JSONPatchType,
							[]byte(`[{"op": "replace", "path": "/spec/kf/config/buildDisableIstioSidecar", "value": false}]`),
							metav1.PatchOptions{},
						)

					// We don't want to fail anything on, just log it.
					if err != nil {
						integration.Logf(t, "failed to revert buildDisableIstioSidecar to false: %v", err)
					}
				})

				// Wait for config-defaults to be updated by the controller.
				for ctx.Err() == nil {
					cm, err := kubeClient.
						CoreV1().
						ConfigMaps("kf").
						Get(ctx, kfconfig.DefaultsConfigName, metav1.GetOptions{})
					testutil.AssertErrorsEqual(t, nil, err)

					defaultCfg, err := kfconfig.NewDefaultsConfigFromConfigMap(cm)
					testutil.AssertErrorsEqual(t, nil, err)

					if defaultCfg.BuildDisableIstioSidecar {
						break
					} else {
						integration.Logf(t, "waiting for %s to be updated...", kfconfig.DefaultsConfigName)
						time.Sleep(5 * time.Second)
					}
				}
			}

			// Push an App to create a Build.
			// NOTE: We can't use CachePush because we want to ensure
			// we get a Build.
			appName := v1alpha1.GenerateName("integration-sans-sidecar", fmt.Sprint(time.Now().UnixNano()))
			// Do this in the background as we really don't care about
			// the App, just the Build.
			go kf.Push(ctx, appName)

			// Look at the pod to ensure it doesn't have anything
			// with Istio.
			istioMatcher := regexp.MustCompile(`^istio`)
			pod := getBuildPod(ctx, t, appName)
			for _, container := range pod.Spec.Containers {
				testutil.AssertFalse(t, "found istio container", istioMatcher.MatchString(container.Name))
			}
		})
	})
}

func kfSystemClient(ctx context.Context) dynamic.ResourceInterface {
	dynamicClient := dynamicclient.Get(ctx)
	return dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Resource: "kfsystems",
	})
}

// TestIntegration_BuildWithResources updates config-defaults to have
// buildPodResources set. It then pushes an app and ensures the resulting
// build has the resources set.
func TestIntegration_BuildWithResources(t *testing.T) {
	integration.RunKubeAPITest(context.Background(), t, func(ctx context.Context, t *testing.T) {
		integration.RunKfTest(ctx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
			kubeClient := kubeclient.Get(ctx)
			kfSystemClient := kfSystemClient(ctx)

			// Patch Config Defaults.
			{
				integration.Logf(t, "patching config-defaults to have buildPodResources set...")
				_, err := kfSystemClient.
					Patch(
						ctx,
						"kfsystem",
						types.JSONPatchType,
						// Pick a weird limit so that we know a default wasn't
						// applied.
						[]byte(`[{"op": "replace", "path": "/spec/kf/config/buildPodResources", "value": {"limits": {"memory": "234Mi"}}}]`),
						metav1.PatchOptions{},
					)

				testutil.AssertErrorsEqual(t, nil, err)

				t.Cleanup(func() {
					integration.Logf(t, "setting cluster's buildPodResources to empty...")
					_, err := kfSystemClient.
						Patch(
							// It's likely the ctx is done.
							context.Background(),
							"kfsystem",
							types.JSONPatchType,
							[]byte(`[{"op": "remove", "path": "/spec/kf/config/buildPodResources"}]`),
							metav1.PatchOptions{},
						)

					// We don't want to fail anything on, just log it.
					if err != nil {
						integration.Logf(t, "failed to revert buildPodResources to empty: %v", err)
					}
				})

				// Wait for config-defaults to be updated by the controller.
				for ctx.Err() == nil {
					cm, err := kubeClient.
						CoreV1().
						ConfigMaps("kf").
						Get(ctx, kfconfig.DefaultsConfigName, metav1.GetOptions{})
					testutil.AssertErrorsEqual(t, nil, err)

					defaultCfg, err := kfconfig.NewDefaultsConfigFromConfigMap(cm)
					testutil.AssertErrorsEqual(t, nil, err)

					if defaultCfg.BuildPodResources != nil && len(defaultCfg.BuildPodResources.Limits) > 0 {
						break
					} else {
						integration.Logf(t, "waiting for %s to be updated...", kfconfig.DefaultsConfigName)
						time.Sleep(5 * time.Second)
					}
				}
			}

			testCases := []struct {
				name                    string
				push                    func(ctx context.Context, appName string)
				containersWithResources []string
			}{
				{
					name: "v2 buildpack",
					push: func(ctx context.Context, appName string) {
						kf.Push(ctx, appName, "--stack=cflinuxfs3")
					},
					containersWithResources: []string{"step-run-lifecycle", "step-build"},
				},
				{
					name: "v3 buildpack",
					push: func(ctx context.Context, appName string) {
						kf.Push(ctx, appName, "--stack=org.cloudfoundry.stacks.cflinuxfs3")
					},
					containersWithResources: []string{"step-build"},
				},
				{
					name: "dockerfile",
					push: func(ctx context.Context, appName string) {
						appPath := filepath.Join(integration.RootDir(ctx, t), "samples", "apps", "helloworld")
						kf.Push(ctx, appName, "--path", appPath, "--dockerfile", "Dockerfile")
					},
					containersWithResources: []string{"step-build"},
				},
			}

			outerCtx := ctx
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					ctx, cancel := context.WithCancel(outerCtx)
					t.Cleanup(cancel)

					// Push an App to create a Build.
					// NOTE: We can't use CachePush because we want to ensure
					// we get a Build.
					appName := v1alpha1.GenerateName("integration-sans-sidecar", fmt.Sprint(time.Now().UnixNano()))
					// Do this in the background as we really don't care about
					// the App, just the Build.
					go tc.push(ctx, appName)

					// Look at the pod to ensure it has the expected resource
					// limit.
					pod := getBuildPod(ctx, t, appName)
					for _, container := range pod.Spec.Containers {
						for _, expectedContainer := range tc.containersWithResources {
							if container.Name != expectedContainer {
								continue
							}

							testutil.AssertEqual(t, "resource memory", container.Resources.Limits[corev1.ResourceMemory], resource.MustParse("234Mi"))
						}
					}
				})
			}
		})
	})
}

func TestIntegration_BuildNodeSelectors(t *testing.T) {
	// b/230261963
	t.Skip()
	integration.RunKubeAPITest(context.Background(), t, func(ctx context.Context, t *testing.T) {
		integration.RunKfTest(ctx, t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
			kubeClient := kubeclient.Get(ctx)
			kfSystemClient := kfSystemClient(ctx)

			// Patch Config Defaults.
			{
				integration.Logf(t, "patching config-defaults to have build node selectors...")
				_, err := kfSystemClient.
					Patch(
						ctx,
						"kfsystem",
						types.JSONPatchType,
						[]byte(`[{"op": "replace", "path": "/spec/kf/config/buildNodeSelectors", "value": {"disktype": "ssd"}}]`),
						metav1.PatchOptions{},
					)

				testutil.AssertErrorsEqual(t, nil, err)

				t.Cleanup(func() {
					integration.Logf(t, "setting cluster's buildPodResources to empty...")
					_, err := kfSystemClient.
						Patch(
							context.Background(),
							"kfsystem",
							types.JSONPatchType,
							[]byte(`[{"op": "remove", "path": "/spec/kf/config/buildNodeSelectors"}]`),
							metav1.PatchOptions{},
						)

					// We don't want to fail anything on, just log it.
					if err != nil {
						integration.Logf(t, "failed to revert buildNodeSelectors to empty: %v", err)
					}
				})

				// Wait for config-defaults to be updated by the controller.
				for ctx.Err() == nil {
					cm, err := kubeClient.
						CoreV1().
						ConfigMaps("kf").
						Get(ctx, kfconfig.DefaultsConfigName, metav1.GetOptions{})
					testutil.AssertErrorsEqual(t, nil, err)

					defaultCfg, err := kfconfig.NewDefaultsConfigFromConfigMap(cm)
					testutil.AssertErrorsEqual(t, nil, err)

					if defaultCfg.BuildPodResources != nil && len(defaultCfg.BuildPodResources.Limits) > 0 {
						break
					} else {
						integration.Logf(t, "waiting for %s to be updated...", kfconfig.DefaultsConfigName)
						time.Sleep(5 * time.Second)
					}
				}
			}

			testCases := []struct {
				name                  string
				push                  func(ctx context.Context, appName string)
				expectedNodeSelectors map[string]string
			}{
				{
					name: "build node selectors are added",
					push: func(ctx context.Context, appName string) {
						kf.Push(ctx, appName)
					},
					expectedNodeSelectors: map[string]string{"disktype": "ssd"},
				},
			}

			outerCtx := ctx
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					ctx, cancel := context.WithCancel(outerCtx)
					t.Cleanup(cancel)

					// Push an App to create a Build.
					// NOTE: We can't use CachePush because we want to ensure
					// we get a Build.
					appName := v1alpha1.GenerateName("integration-buildNodeSelectors", fmt.Sprint(time.Now().UnixNano()))
					// Do this in the background as we really don't care about
					// the App, just the Build.
					go tc.push(ctx, appName)

					// Look at the pod to ensure it has the expected resource
					// limit.
					pod := getBuildPod(ctx, t, appName)

					// As long as the node selectors appear on the build pods, Kubernetes makes sure the build pod is assigned to the right nodes.
					testutil.AssertEqual(t, "node selectors", tc.expectedNodeSelectors, pod.Spec.NodeSelector)
				})
			}
		})
	})
}

func getBuildPod(ctx context.Context, t *testing.T, appName string) *corev1.Pod {
	t.Helper()
	var pod *corev1.Pod
	kfClient := kfclient.Get(ctx).KfV1alpha1()
	tknClient := tektonclient.Get(ctx).TektonV1beta1()
	kubeClient := kubeclient.Get(ctx)

	// App -> TaskRun -> Pod
	spaceName := integration.SpaceFromContext(ctx)

	// Wait for the BuildName to be populated.
	var buildName string
	{
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		t.Cleanup(cancel)
		for ctx.Err() == nil {
			app, err := kfClient.
				Apps(spaceName).
				Get(ctx, appName, metav1.GetOptions{})

			if apierrs.IsNotFound(err) {
				integration.Logf(t, "App %s not found...", appName)
				time.Sleep(time.Second)
				continue
			} else if err != nil {
				t.Fatal(err)
			} else if app.Status.LatestCreatedBuildName == "" {
				integration.Logf(t, "App %s doesn't have the status.latestBuild populated yet...", appName)
				time.Sleep(time.Second)
				continue
			}

			buildName = app.Status.LatestCreatedBuildName
			break
		}
	}

	// Wait for the PodName to be populated.
	var podName string
	{
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		t.Cleanup(cancel)
		for ctx.Err() == nil {
			tr, err := tknClient.
				TaskRuns(spaceName).
				Get(ctx, buildName, metav1.GetOptions{})

			if apierrs.IsNotFound(err) {
				integration.Logf(t, "TaskRun %s not found...", buildName)
				time.Sleep(time.Second)
				continue
			} else if err != nil {
				t.Fatal(err)
			} else if tr.Status.PodName == "" {
				integration.Logf(t, "TaskRun %s doesn't have the status.podName populated yet...", appName)
				time.Sleep(time.Second)
				continue
			}

			podName = tr.Status.PodName
			break
		}
	}

	pod, err := kubeClient.
		CoreV1().
		Pods(spaceName).
		Get(ctx, podName, metav1.GetOptions{})
	testutil.AssertErrorsEqual(t, nil, err)

	return pod
}
