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

package servicebindings

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	cfutil "github.com/google/kf/pkg/kf/cfutil"
	. "github.com/google/kf/pkg/kf/testutil"
)

func TestIntegration_Marketplace(t *testing.T) {
	t.Skip()
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		withServiceBroker(ctx, t, kf, func(ctx context.Context) {
			marketplaceOutput := kf.Marketplace(ctx)
			AssertContainsAll(t, strings.Join(marketplaceOutput, "\n"), []string{BrokerFromContext(ctx), "Active"})
		})
	})
}

func TestIntegration_Services(t *testing.T) {
	t.Skip()
	checkClusterStatus(t)
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		withServiceBroker(ctx, t, kf, func(ctx context.Context) {
			withServiceInstance(ctx, kf, func(ctx context.Context) {
				servicesOutput := kf.Services(ctx)
				AssertContainsAll(t, strings.Join(servicesOutput, "\n"), []string{ServiceInstanceFromContext(ctx),
					ServiceClassFromContext(ctx), ServicePlanFromContext(ctx), "ProvisionedSuccessfully"})
			})
		})
	})
}

func TestIntegration_Bindings(t *testing.T) {
	t.Skip()
	checkClusterStatus(t)
	appName := fmt.Sprintf("integration-binding-app-%d", time.Now().UnixNano())
	appPath := "./samples/apps/envs"
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		withServiceBroker(ctx, t, kf, func(ctx context.Context) {
			withServiceInstance(ctx, kf, func(ctx context.Context) {
				withApp(ctx, t, kf, appName, appPath, false, func(ctx context.Context) {
					withServiceBinding(ctx, t, kf, func(ctx context.Context) {
						bindingsOutput := kf.Bindings(ctx)
						AssertContainsAll(t, strings.Join(bindingsOutput, "\n"), []string{AppFromContext(ctx),
							ServiceInstanceFromContext(ctx), "True", "InjectedBindResult"})
					})
				})
			})
		})
	})
}

func TestIntegration_VcapServices(t *testing.T) {
	t.Skip("re-enable me when namespaced brokers are added")
	checkClusterStatus(t)
	appName := fmt.Sprintf("integration-binding-app-%d", time.Now().UnixNano())
	appPath := "./samples/apps/envs"
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		withServiceBroker(ctx, t, kf, func(ctx context.Context) {
			withServiceInstance(ctx, kf, func(ctx context.Context) {
				withApp(ctx, t, kf, appName, appPath, false, func(ctx context.Context) {
					// Assert VCAP_SERVICES is blank
					vcs := extractVcapServices(ctx, t, kf)
					AssertEqual(t, "bound service count", 0, len(vcs))

					withServiceBinding(ctx, t, kf, func(ctx context.Context) {
						vcs := extractVcapServices(ctx, t, kf)

						expected := cfutil.VcapServicesMap{
							ServiceClassFromContext(ctx): []cfutil.VcapService{
								{
									BindingName:  ServiceInstanceFromContext(ctx),
									InstanceName: ServiceInstanceFromContext(ctx),
									Name:         ServiceInstanceFromContext(ctx),
									Label:        ServiceClassFromContext(ctx),
									Tags:         []string(nil),
									Plan:         ServicePlanFromContext(ctx),
									Credentials: map[string]string{
										"password": "fake-pw",
										"username": "fake-user",
									},
								},
							},
						}

						AssertEqual(t, "vcap services", expected, vcs)
					})
				})
			})
		})
	})
}

func TestIntegration_VcapServices_customBindingName(t *testing.T) {
	checkClusterStatus(t)
	appName := fmt.Sprintf("integration-binding-app-%d", time.Now().UnixNano())
	appPath := "./samples/apps/envs"
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		withServiceBroker(ctx, t, kf, func(ctx context.Context) {
			withServiceInstance(ctx, kf, func(ctx context.Context) {
				withApp(ctx, t, kf, appName, appPath, false, func(ctx context.Context) {
					serviceInstanceName := ServiceInstanceFromContext(ctx)
					appName := AppFromContext(ctx)
					kf.BindService(ctx, appName, serviceInstanceName, "--binding-name", "my-binding")
					defer kf.UnbindService(ctx, appName, serviceInstanceName) // cleanup

					vcs := extractVcapServices(ctx, t, kf)

					expected := cfutil.VcapServicesMap{
						ServiceClassFromContext(ctx): []cfutil.VcapService{
							{
								BindingName:  "my-binding",
								InstanceName: ServiceInstanceFromContext(ctx),
								Name:         "my-binding",
								Label:        ServiceClassFromContext(ctx),
								Tags:         []string(nil),
								Plan:         ServicePlanFromContext(ctx),
								Credentials: map[string]string{
									"password": "fake-pw",
									"username": "fake-user",
								},
							},
						},
					}

					AssertEqual(t, "vcap services", expected, vcs)
				})
			})
		})
	})
}

var checkOnce sync.Once

func checkClusterStatus(t *testing.T) {
	checkOnce.Do(func() {
		testIntegration_Doctor(t)
	})
}

func extractVcapServices(ctx context.Context, t *testing.T, kf *Kf) cfutil.VcapServicesMap {
	vcapServicesOutput := kf.VcapServices(ctx, AppFromContext(ctx))
	vcapServices := cfutil.VcapServicesMap{}
	if err := json.Unmarshal([]byte(strings.Join(vcapServicesOutput, "\n")), &vcapServices); err != nil {
		t.Fatalf("couldn't unmarshal VCAP Services: %s", err)
	}

	return vcapServices
}

// testIntegration_Doctor runs the doctor command. It ensures the cluster the
// tests are running against is in good shape.
func testIntegration_Doctor(t *testing.T) {
	RunKfTest(t, func(ctx context.Context, t *testing.T, kf *Kf) {
		kf.Doctor(ctx)
	})
}

func withServiceBroker(ctx context.Context, t *testing.T, kf *Kf, callback func(newCtx context.Context)) {
	brokerAppName := fmt.Sprintf("integration-broker-app-%d", time.Now().UnixNano())
	brokerPath := "./samples/apps/service-broker"
	brokerName := fmt.Sprintf("fake-broker-%d", time.Now().UnixNano())

	withApp(ctx, t, kf, brokerAppName, brokerPath, true, func(ctx context.Context) {
		// Register the mock service broker to service catalog, and then clean it up.
		kf.CreateServiceBroker(ctx, brokerName, internalBrokerUrl(brokerAppName, SpaceFromContext(ctx)))

		// Temporary solution to allow service broker registration to complete.
		// TODO: Add flag to run the command synchronously.
		time.Sleep(2 * time.Second)
		defer kf.DeleteServiceBroker(ctx, brokerName)

		ctx = ContextWithBroker(ctx, brokerName)
		callback(ctx)
	})
}

func withServiceInstance(ctx context.Context, kf *Kf, callback func(newCtx context.Context)) {
	serviceClass := "fake-service" // service class provided by the mock broker
	servicePlan := "fake-plan"     // service plan provided by the mock broker
	serviceInstanceName := "int-service-instance"

	kf.CreateService(ctx, serviceClass, servicePlan, serviceInstanceName)
	defer kf.DeleteService(ctx, serviceInstanceName)

	ctx = ContextWithServiceClass(ctx, serviceClass)
	ctx = ContextWithServicePlan(ctx, servicePlan)
	ctx = ContextWithServiceInstance(ctx, serviceInstanceName)
	callback(ctx)
}

func withServiceBinding(ctx context.Context, t *testing.T, kf *Kf, callback func(newCtx context.Context)) {
	serviceInstanceName := ServiceInstanceFromContext(ctx)
	appName := AppFromContext(ctx)
	kf.BindService(ctx, appName, serviceInstanceName)
	defer kf.UnbindService(ctx, appName, serviceInstanceName)

	callback(ctx)
}

func withApp(ctx context.Context, t *testing.T, kf *Kf, appName string, path string, isBroker bool, callback func(newCtx context.Context)) {
	// Push the app then clean it up.
	kf.Push(ctx, appName,
		"--path", filepath.Join(RootDir(ctx, t), path),
	)
	defer kf.Delete(ctx, appName)

	if !isBroker {
		ctx = ContextWithApp(ctx, appName)
	}

	callback(ctx)
}

func internalBrokerUrl(brokerName string, namespace string) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local", brokerName, namespace)
}
