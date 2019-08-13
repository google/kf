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
)

func TestXxx(t *testing.T) {
}

/*
func TestClient_GetVcapServices(t *testing.T) {
	fakeInstance := apiv1beta1.ServiceInstance{}
	fakeInstance.Name = "my-instance"

	fakeBinding := apiv1beta1.ServiceBinding{}
	fakeBinding.Name = "my-binding"
	fakeBinding.Labels = map[string]string{
		servicebindings.AppNameLabel:     "my-app",
		servicebindings.BindingNameLabel: "binding-name",
	}
	fakeBinding.Spec.SecretName = "my-secret"
	fakeBinding.Spec.InstanceRef.Name = fakeInstance.Name

	fakeBindingList := &apiv1beta1.ServiceBindingList{
		Items: []apiv1beta1.ServiceBinding{fakeBinding},
	}

	emptyBindingList := &apiv1beta1.ServiceBindingList{
		Items: []apiv1beta1.ServiceBinding{},
	}

	fakeSecret := corev1.Secret{}
	fakeSecret.Name = "my-secret"
	fakeSecret.Data = map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
	}

	cases := map[string]ServiceBindingApiTestCase{
		"api-error": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("api-error"))

				_, err := client.GetVcapServices("my-app")
				testutil.AssertErrorsEqual(t, errors.New("api-error"), err)
			},
		},
		"default options": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(emptyBindingList, nil)

				_, err := client.GetVcapServices("my-app")
				testutil.AssertNil(t, "GetVcapServices err", err)
			},
		},
		"gets secret": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(fakeBindingList, nil)

				deps.apiserver.EXPECT().
					Get(gomock.Any(), "default", "my-instance").
					Return(&fakeInstance, nil)

				deps.secrets.EXPECT().Get("my-secret", gomock.Any()).Return(&fakeSecret, nil)

				actualVcap, err := client.GetVcapServices("my-app")
				testutil.AssertNil(t, "GetVcapServices err", err)

				expectedVcap := servicebindings.VcapServicesMap{}
				expectedVcap.Add(servicebindings.NewVcapService(fakeInstance, fakeBinding, &fakeSecret))
				testutil.AssertEqual(t, "vcap services", expectedVcap, actualVcap)
			},
		},
		"fail on bad secret": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {
				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(fakeBindingList, nil)

				deps.apiserver.EXPECT().
					Get(gomock.Any(), "default", "my-instance").
					Return(&fakeInstance, nil)

				deps.secrets.EXPECT().
					Get("my-secret", gomock.Any()).
					Return(nil, errors.New("secret doesn't exist"))

				_, actualErr := client.GetVcapServices("my-app", servicebindings.WithGetVcapServicesFailOnBadSecret(true))
				expectedErr := errors.New("couldn't create VCAP_SERVICES, the secret for binding my-binding couldn't be fetched: secret doesn't exist")
				testutil.AssertErrorsEqual(t, expectedErr, actualErr)
			},
		},
		"no fail on bad secret": {
			Run: func(t *testing.T, deps fakeDependencies, client servicebindings.ClientInterface) {

				deps.apiserver.EXPECT().
					List(gomock.Any(), "default", gomock.Any(), gomock.Any()).
					Return(fakeBindingList, nil)

				deps.apiserver.EXPECT().
					Get(gomock.Any(), "default", "my-instance").
					Return(&fakeInstance, nil)

				deps.secrets.EXPECT().Get("my-secret", gomock.Any()).Return(nil, errors.New("secret doesn't exist"))

				// VCAP should be empty
				actualVcap, err := client.GetVcapServices("my-app")
				testutil.AssertNil(t, "GetVcapServices err", err)

				expectedVcap := servicebindings.VcapServicesMap{}
				testutil.AssertEqual(t, "vcap services", expectedVcap, actualVcap)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, tc.ExecuteTest)
	}
}
*/
