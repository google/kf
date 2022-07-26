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
	"os"
	"strings"
	"testing"
	"time"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/cfutil"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/google/kf/v2/pkg/kf/testutil/integration"
	"github.com/google/kf/v2/pkg/reconciler/app/resources"
)

const (
	// These username and password values match the values defined in the fake
	// service broker App.
	validBrokerUsername = "valid-user"
	validBrokerPassword = "valid-pw"

	// selfSignedCertsBrokerURLEnv is the envvar that when set, turns on the
	// self signed certs test. It will look for the
	// https://github.com/cloudfoundry-community/worlds-simplest-service-broker
	// broker with the username and password of 'broker'.
	selfSignedCertsBrokerURLEnv = "SELF_SIGNED_CERTS_BROKER_URL"
)

func TestIntegration_Marketplace_selfSignedCerts(t *testing.T) {
	t.Skip("This test wasn't designed to run hermetically, it should be updated.")

	u := os.Getenv(selfSignedCertsBrokerURLEnv)
	if u == "" {
		t.Skipf("env %q not set, skipping...", selfSignedCertsBrokerURLEnv)
		return
	}

	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		brokerName := fmt.Sprintf("self-signed-cert-broker-%d", time.Now().UnixNano())

		// Simply the act of creating the service broker will prove if the
		// feature works or not. If this is not working correctly, the
		// create-service-broker command will fail.
		kf.CreateServiceBroker(ctx, brokerName, "broker", "broker", u)

		// Cleanup the service broker.
		kf.DeleteServiceBroker(ctx, brokerName)
	})
}

func TestIntegration_Marketplace(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		withServiceBroker(ctx, t, kf, func(ctx context.Context) {
			marketplaceOutput := kf.Marketplace(ctx)
			testutil.AssertContainsAll(t, strings.Join(marketplaceOutput, "\n"), []string{integration.BrokerFromContext(ctx)})
		})
	})
}

func TestIntegration_Services(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		withServiceBroker(ctx, t, kf, func(ctx context.Context) {
			withServiceInstance(ctx, kf, func(ctx context.Context) {
				servicesOutput := kf.Services(ctx)
				testutil.AssertContainsAll(t, strings.Join(servicesOutput, "\n"), []string{
					integration.ServiceInstanceFromContext(ctx),
					integration.ServiceClassFromContext(ctx),
					integration.ServicePlanFromContext(ctx),
					"Ready",
				})
			})
		})
	})
}

func TestIntegration_VolumeServices(t *testing.T) {
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		withVolumeServiceInstance(ctx, kf, func(ctx context.Context) {
			servicesOutput := kf.Services(ctx)
			testutil.AssertContainsAll(t, strings.Join(servicesOutput, "\n"), []string{
				integration.ServiceInstanceFromContext(ctx),
				integration.ServiceClassFromContext(ctx),
				integration.ServicePlanFromContext(ctx),
				"Ready",
			})
		})
	})
}

func TestIntegration_Bindings(t *testing.T) {
	appName := fmt.Sprintf("integration-binding-app-%d", time.Now().UnixNano())
	appPath := "./samples/apps/envs"
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		withServiceBroker(ctx, t, kf, func(ctx context.Context) {
			withServiceInstance(ctx, kf, func(ctx context.Context) {
				integration.WithApp(ctx, t, kf, appName, appPath, false, func(ctx context.Context) {
					withServiceBinding(ctx, kf, func(ctx context.Context) {
						bindingsOutput := kf.Bindings(ctx)
						testutil.AssertContainsAll(t, strings.Join(bindingsOutput, "\n"), []string{
							integration.AppFromContext(ctx),
							integration.ServiceInstanceFromContext(ctx),
						})
					})
				})
			})
		})
	})
}

func TestIntegration_VcapServices_customBindingName(t *testing.T) {
	appName := fmt.Sprintf("integration-binding-app-%d", time.Now().UnixNano())
	bindingName := fmt.Sprintf("bind-%x", time.Now().UnixNano())
	appPath := "./samples/apps/envs"
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		withServiceBroker(ctx, t, kf, func(ctx context.Context) {
			withServiceInstance(ctx, kf, func(ctx context.Context) {
				integration.WithApp(ctx, t, kf, appName, appPath, false, func(ctx context.Context) {
					serviceInstanceName := integration.ServiceInstanceFromContext(ctx)
					appName := integration.AppFromContext(ctx)
					kf.BindService(ctx, appName, serviceInstanceName, "--binding-name", bindingName)
					defer kf.UnbindService(ctx, appName, serviceInstanceName) // cleanup

					vcs := extractVcapServices(ctx, t, kf)
					fakePw := json.RawMessage(`"fake-pw"`)
					fakeUser := json.RawMessage(`"fake-user"`)
					someArray := json.RawMessage(`["a",2]`)
					someInt := json.RawMessage("1")
					someBool := json.RawMessage("true")
					quotes := json.RawMessage(`"\"quoted\""`)
					quotesInside := json.RawMessage(`"abc\"d"`)
					expected := cfutil.VcapServicesMap{
						integration.ServiceClassFromContext(ctx): []cfutil.VcapService{
							{
								BindingName:  &bindingName,
								InstanceName: integration.ServiceInstanceFromContext(ctx),
								Name:         bindingName,
								Label:        integration.ServiceClassFromContext(ctx),
								Tags:         []string{"fake-tag"},
								Plan:         integration.ServicePlanFromContext(ctx),
								Credentials: map[string]json.RawMessage{
									"password":     fakePw,
									"username":     fakeUser,
									"somearray":    someArray,
									"someint":      someInt,
									"somebool":     someBool,
									"quotes":       quotes,
									"quotesinside": quotesInside,
								},
							},
						},
					}

					testutil.AssertEqual(t, "vcap services", expected, vcs)
				})
			})
		})
	})
}

func TestIntegration_VcapServices_userProvidedService(t *testing.T) {
	appName := fmt.Sprintf("integration-binding-app-%d", time.Now().UnixNano())
	appPath := "./samples/apps/envs"
	integration.RunKfTest(context.Background(), t, func(ctx context.Context, t *testing.T, kf *integration.Kf) {
		withUserProvidedServiceInstance(ctx, kf, func(ctx context.Context) {
			integration.WithApp(ctx, t, kf, appName, appPath, false, func(ctx context.Context) {
				serviceInstanceName := integration.ServiceInstanceFromContext(ctx)
				appName := integration.AppFromContext(ctx)
				kf.BindService(ctx, appName, serviceInstanceName)

				vcs := extractVcapServices(ctx, t, kf)
				testUser := json.RawMessage(`"test-user"`)
				testPw := json.RawMessage(`"test-pw"`)
				someInt := json.RawMessage("5")
				expected := cfutil.VcapServicesMap{
					integration.ServiceClassFromContext(ctx): []cfutil.VcapService{
						{
							InstanceName: integration.ServiceInstanceFromContext(ctx),
							Name:         integration.ServiceInstanceFromContext(ctx),
							Label:        integration.ServiceClassFromContext(ctx),
							Tags:         []string{},
							Credentials: map[string]json.RawMessage{
								"username": testUser,
								"password": testPw,
								"someint":  someInt,
							},
						},
					},
				}

				testutil.AssertEqual(t, "vcap services", expected, vcs)

				// Update user-provided service credentials
				newCredentials := `{"username":"test-user", "password":"new-pw", "newKey":6}`
				kf.UpdateUserProvidedService(ctx, serviceInstanceName, "-p", newCredentials)
				defer kf.UnbindService(ctx, appName, serviceInstanceName) // cleanup

				// Check that credentials are updated in VCAP_SERVICES.
				vcs = extractVcapServices(ctx, t, kf)
				newPw := json.RawMessage(`"new-pw"`)
				newInt := json.RawMessage("6")
				expected = cfutil.VcapServicesMap{
					integration.ServiceClassFromContext(ctx): []cfutil.VcapService{
						{
							InstanceName: integration.ServiceInstanceFromContext(ctx),
							Name:         integration.ServiceInstanceFromContext(ctx),
							Label:        integration.ServiceClassFromContext(ctx),
							Tags:         []string{},
							Credentials: map[string]json.RawMessage{
								"username": testUser,
								"password": newPw,
								"newKey":   newInt,
							},
						},
					},
				}

				testutil.AssertEqual(t, "updated vcap services", expected, vcs)
			})
		})
	})
}

func extractVcapServices(
	ctx context.Context,
	t *testing.T,
	kf *integration.Kf,
) cfutil.VcapServicesMap {
	vcapServicesOutput := kf.VcapServices(ctx, integration.AppFromContext(ctx))
	vcapServices := cfutil.VcapServicesMap{}
	if err := json.Unmarshal(vcapServicesOutput, &vcapServices); err != nil {
		t.Fatalf("couldn't unmarshal VCAP Services: %s", err)
	}

	return vcapServices
}

func withServiceBroker(
	ctx context.Context,
	t *testing.T,
	kf *integration.Kf,
	callback func(newCtx context.Context),
) {
	brokerAppName := fmt.Sprintf("integration-broker-app-%d", time.Now().UnixNano())
	brokerPath := "./samples/apps/service-broker"
	brokerName := fmt.Sprintf("fake-broker-%d", time.Now().UnixNano())

	integration.WithAppArgs(ctx, t, kf, brokerAppName, brokerPath, true, []string{"--health-check-type", "http"}, func(ctx context.Context) {
		// Register the mock service broker to service catalog, and then clean it up.
		kf.CreateServiceBroker(ctx, brokerName, validBrokerUsername, validBrokerPassword, internalBrokerURL(brokerAppName, integration.SpaceFromContext(ctx)), "--space-scoped")

		defer kf.DeleteServiceBroker(ctx, brokerName, "--space-scoped")

		ctx = integration.ContextWithBroker(ctx, brokerName)
		callback(ctx)
	})
}

func withServiceInstance(
	ctx context.Context,
	kf *integration.Kf,
	callback func(newCtx context.Context),
) {
	serviceClass := "fake-service" // service class provided by the mock broker
	servicePlan := "fake-plan"     // service plan provided by the mock broker
	serviceInstanceName := "int-service-instance"
	brokerName := integration.BrokerFromContext(ctx)

	kf.CreateService(ctx, serviceClass, servicePlan, serviceInstanceName, "-b", brokerName)
	defer kf.DeleteService(ctx, serviceInstanceName)

	ctx = integration.ContextWithServiceClass(ctx, serviceClass)
	ctx = integration.ContextWithServicePlan(ctx, servicePlan)
	ctx = integration.ContextWithServiceInstance(ctx, serviceInstanceName)
	callback(ctx)
}

func withUserProvidedServiceInstance(
	ctx context.Context,
	kf *integration.Kf,
	callback func(newCtx context.Context),
) {
	serviceInstanceName := "int-ups-instance"
	credentials := `{"username":"test-user", "password":"test-pw", "someint":5}`
	kf.CreateUserProvidedService(ctx, serviceInstanceName, "-p", credentials)
	defer kf.DeleteService(ctx, serviceInstanceName)

	ctx = integration.ContextWithServiceInstance(ctx, serviceInstanceName)
	ctx = integration.ContextWithServiceClass(ctx, v1alpha1.UserProvidedServiceClassName)
	callback(ctx)
}

func withVolumeServiceInstance(
	ctx context.Context,
	kf *integration.Kf,
	callback func(newCtx context.Context),
) {
	serviceClass := "nfs"
	servicePlan := "existing"
	serviceInstanceName := "volume-service-instance"

	kf.CreateVolumeService(
		ctx,
		serviceInstanceName,
		"-c", "{\"share\":\"host/share\", \"capacity\":\"1Gi\"}",
	)

	defer kf.DeleteService(ctx, serviceInstanceName)

	ctx = integration.ContextWithServiceClass(ctx, serviceClass)
	ctx = integration.ContextWithServicePlan(ctx, servicePlan)
	ctx = integration.ContextWithServiceInstance(ctx, serviceInstanceName)
	callback(ctx)
}

func withServiceBinding(
	ctx context.Context,
	kf *integration.Kf,
	callback func(newCtx context.Context),
) {
	serviceInstanceName := integration.ServiceInstanceFromContext(ctx)
	appName := integration.AppFromContext(ctx)
	kf.BindService(ctx, appName, serviceInstanceName)
	defer kf.UnbindService(ctx, appName, serviceInstanceName)

	callback(ctx)
}

func internalBrokerURL(brokerName string, namespace string) string {
	// User the URL provided by the service on the App via kube DNS.
	return fmt.Sprintf("http://%s.%s.svc", resources.ServiceNameForAppName(brokerName), namespace)
}
