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

package resources

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeRouteServiceURL(scheme, host string, path string) *v1alpha1.RouteServiceURL {
	return &v1alpha1.RouteServiceURL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}
}

func makeRouteServiceURLWithPort(scheme, host string, port int32, path string) *v1alpha1.RouteServiceURL {
	rsURL := makeRouteServiceURL(scheme, host, path)
	hostWithPort := fmt.Sprintf("%s:%d", rsURL.Host, port)
	rsURL.Host = hostWithPort
	return rsURL
}

func TestMakeDeployment(t *testing.T) {
	happyRouteServiceFields := makeRouteServiceURLWithPort("http", "auth.my-route-svc.com", 80, "/some-path")
	happyServiceInstance := &v1alpha1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-route-svc",
		},
		Spec: v1alpha1.ServiceInstanceSpec{
			ServiceType: v1alpha1.ServiceType{
				UPS: &v1alpha1.UPSInstance{
					RouteServiceURL: happyRouteServiceFields,
				},
			},
		},
	}
	for tn, tc := range map[string]struct {
		serviceInstance *v1alpha1.ServiceInstance
		cfg             *config.Config
		wantErr         error
	}{
		"missing image in config": {
			serviceInstance: happyServiceInstance,
			cfg:             config.CreateConfigForTest(&config.DefaultsConfig{}),
			wantErr:         errors.New("config value for RouteServiceProxyImage couldn't be found"),
		},
		"happy": {
			serviceInstance: happyServiceInstance,
			cfg: config.CreateConfigForTest(&config.DefaultsConfig{
				RouteServiceProxyImage: "gcr.io/fake/proxy/image",
			}),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			actualDeployment, actualErr := MakeDeployment(tc.serviceInstance, tc.cfg)
			testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
			configDefaults, err := tc.cfg.Defaults()
			testutil.AssertNil(t, "err", err)
			testutil.AssertGoldenJSONContext(t, "deployment", actualDeployment, map[string]interface{}{
				"serviceInstance": tc.serviceInstance,
				"config.defaults": configDefaults,
			})
		})
	}
}
