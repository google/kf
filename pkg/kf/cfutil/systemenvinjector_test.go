// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cfutil_test

import (
	"testing"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	servicecatalogclient "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned/fake"
	"github.com/google/kf/pkg/kf/cfutil"
	"github.com/google/kf/pkg/kf/testutil"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var (
	app = &v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-app",
		},
	}

	serviceInstance = &servicecatalogv1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-instance",
		},
		Spec: servicecatalogv1beta1.ServiceInstanceSpec{
			PlanReference: servicecatalogv1beta1.PlanReference{
				ClusterServiceClassExternalName: "my-class",
				ClusterServicePlanExternalName:  "my-plan",
			},
		},
	}

	serviceBinding = &servicecatalogv1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-binding",
			Labels: map[string]string{
				"kf-binding-name": "my-binding-name",
			},
		},
		Spec: servicecatalogv1beta1.ServiceBindingSpec{
			InstanceRef: servicecatalogv1beta1.LocalObjectReference{
				Name: "my-instance",
			},
		},
	}

	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "binding-secret",
		},
	}
)

func Test_GetVcapServices(t *testing.T) {

	servicecatalogClient := servicecatalogclient.NewSimpleClientset(serviceInstance)
	k8sClient := k8sfake.NewSimpleClientset(secret)

	systemEnvInjector := cfutil.NewSystemEnvInjector(servicecatalogClient, k8sClient)

	cases := map[string]struct {
		Run func(t *testing.T, systemEnvInjector cfutil.SystemEnvInjector)
	}{
		"happy": {
			Run: func(t *testing.T, systemEnvInjector cfutil.SystemEnvInjector) {

				vcapService, err := systemEnvInjector.GetVcapService(app.Name, serviceBinding)
				testutil.AssertNil(t, "error", err)
				testutil.AssertEqual(t, "name", "my-binding", vcapService.Name)
				testutil.AssertEqual(t, "instance name", "my-instance", vcapService.InstanceName)
				testutil.AssertEqual(t, "label", "my-class", vcapService.Label)
				testutil.AssertEqual(t, "tags", []string{}, vcapService.Tags)
				testutil.AssertEqual(t, "plan", "my-plan", vcapService.Plan)
				testutil.AssertEqual(t, "credentials", map[string]string{}, vcapService.Credentials)
				testutil.AssertEqual(t, "binding name", "my-binding-name", vcapService.BindingName)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) { tc.Run(t, systemEnvInjector) })
	}
}

func TestSystemEnvInjector(t *testing.T) {
	t.Parallel()

	servicecatalogClient := servicecatalogclient.NewSimpleClientset(serviceInstance)
	k8sClient := k8sfake.NewSimpleClientset(secret)

	systemEnvInjector := cfutil.NewSystemEnvInjector(servicecatalogClient, k8sClient)

	cases := map[string]struct {
		Run func(t *testing.T, systemEnvInjector cfutil.SystemEnvInjector)
	}{
		"happy": {
			Run: func(t *testing.T, systemEnvInjector cfutil.SystemEnvInjector) {

				env, err := systemEnvInjector.ComputeSystemEnv(app, []servicecatalogv1beta1.ServiceBinding{*serviceBinding})
				testutil.AssertNil(t, "error", err)
				testutil.AssertEqual(t, "env count", 2, len(env))
				hasVcapApplication := false
				hasVcapServices := false
				for _, envVar := range env {
					if envVar.Name == "VCAP_APPLICATION" {
						hasVcapApplication = true
					}
					if envVar.Name == "VCAP_SERVICES" {
						hasVcapServices = true
					}

				}

				if !hasVcapServices {
					t.Fatal("Expected map to contain VCAP_SERVICES")
				}

				if !hasVcapApplication {
					t.Fatal("Expected map to contain VCAP_APPLICATION")
				}
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) { tc.Run(t, systemEnvInjector) })
	}
}
