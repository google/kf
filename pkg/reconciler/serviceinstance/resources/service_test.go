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
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMakeService(t *testing.T) {
	happyRouteServiceURL := makeRouteServiceURLWithPort("http", "auth.my-route-svc.com", 80, "/some-path")
	happyServiceInstance := &v1alpha1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-route-svc",
		},
		Spec: v1alpha1.ServiceInstanceSpec{
			ServiceType: v1alpha1.ServiceType{
				UPS: &v1alpha1.UPSInstance{
					RouteServiceURL: happyRouteServiceURL,
				},
			},
		},
	}

	for tn, tc := range map[string]struct {
		serviceInstance *v1alpha1.ServiceInstance
	}{
		"default": {
			serviceInstance: happyServiceInstance,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			svc := MakeService(tc.serviceInstance)
			testutil.AssertGoldenJSONContext(t, "service", svc, map[string]interface{}{
				"serviceInstance": tc.serviceInstance,
			})
		})
	}
}
