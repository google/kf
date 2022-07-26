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
	"context"
	"encoding/json"
	"testing"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/cfutil"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

var (
	appName               = "my-app"
	instanceName          = "my-instance"
	bindingGeneratedName  = "my-binding-generated-name"
	customBindingName     = "my-custom-binding"
	credentialsSecretName = "binding-secret"
	tags                  = []string{"tag1", "tag2"}
	className             = "my-class"
	planName              = "my-plan"
	credsMap              = map[string][]byte{
		"uri": json.RawMessage(`"postgres://user:pass@mydbinstance:5432/mydb"`),
	}
	app = &v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
	}

	credentialsSecret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: credentialsSecretName,
		},
		Data: credsMap,
	}

	serviceBinding = v1alpha1.ServiceInstanceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: bindingGeneratedName,
		},
		Spec: v1alpha1.ServiceInstanceBindingSpec{
			BindingType: v1alpha1.BindingType{
				App: &v1alpha1.AppRef{
					Name: appName,
				},
			},
			InstanceRef: corev1.LocalObjectReference{
				Name: instanceName,
			},
			BindingNameOverride: customBindingName,
		},
		Status: v1alpha1.ServiceInstanceBindingStatus{
			BindingName: customBindingName,
			CredentialsSecretRef: corev1.LocalObjectReference{
				Name: credentialsSecretName,
			},
			ServiceFields: v1alpha1.ServiceFields{
				Tags:      tags,
				ClassName: className,
				PlanName:  planName,
			},
		},
	}
)

func Test_GetVcapServices(t *testing.T) {
	k8sClient := k8sfake.NewSimpleClientset(credentialsSecret)

	systemEnvInjector := cfutil.NewSystemEnvInjector(k8sClient)

	cases := map[string]struct {
		Run func(t *testing.T, systemEnvInjector cfutil.SystemEnvInjector)
	}{
		"happy": {
			Run: func(t *testing.T, systemEnvInjector cfutil.SystemEnvInjector) {

				vcapService, err := systemEnvInjector.GetVcapService(context.Background(), app.Name, serviceBinding)
				testutil.AssertNil(t, "error", err)
				testutil.AssertEqual(t, "name", customBindingName, vcapService.Name)
				testutil.AssertEqual(t, "instance name", instanceName, vcapService.InstanceName)
				testutil.AssertEqual(t, "label", className, vcapService.Label)
				testutil.AssertEqual(t, "tags", tags, vcapService.Tags)
				testutil.AssertEqual(t, "plan", planName, vcapService.Plan)
				testutil.AssertEqual(t, "credentials", map[string]json.RawMessage{
					"uri": json.RawMessage(`"postgres://user:pass@mydbinstance:5432/mydb"`),
				}, vcapService.Credentials)
				testutil.AssertEqual(t, "binding name", customBindingName, *vcapService.BindingName)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) { tc.Run(t, systemEnvInjector) })
	}
}

func TestSystemEnvInjector(t *testing.T) {
	t.Parallel()

	k8sClient := k8sfake.NewSimpleClientset(credentialsSecret)

	systemEnvInjector := cfutil.NewSystemEnvInjector(k8sClient)

	cases := map[string]struct {
		Run func(t *testing.T, systemEnvInjector cfutil.SystemEnvInjector)
	}{
		"happy": {
			Run: func(t *testing.T, systemEnvInjector cfutil.SystemEnvInjector) {

				env, err := systemEnvInjector.ComputeSystemEnv(context.Background(), app, []v1alpha1.ServiceInstanceBinding{serviceBinding})
				testutil.AssertNil(t, "error", err)
				testutil.AssertEqual(t, "env count", 2, len(env))

				hasVcapServices := false
				hasDbURL := false

				for _, envVar := range env {
					if envVar.Name == cfutil.VcapServicesEnvVarName {
						hasVcapServices = true
					}
					if envVar.Name == cfutil.DatabaseURLEnvVarName {
						hasDbURL = true
						testutil.AssertEqual(t, "db url", "postgres://user:pass@mydbinstance:5432/mydb", envVar.Value)
					}
				}

				if !hasVcapServices {
					t.Fatal("Expected map to contain VCAP_SERVICES")
				}

				if !hasDbURL {
					t.Fatal("Expected map to contain DATABASE_URL")
				}
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) { tc.Run(t, systemEnvInjector) })
	}
}
