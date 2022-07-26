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

package kfvalidation

import (
	"context"
	"errors"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kffake "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/fake"
	"github.com/google/kf/v2/pkg/client/kf/informers/externalversions"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"k8s.io/client-go/tools/cache"
)

func TestValidateServiceInstanceNotExists(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	brokeredInstance := &v1alpha1.ServiceInstance{}
	brokeredInstance.Name = "brokered-instance"
	brokeredInstance.Namespace = "test-ns"
	brokeredInstance.APIVersion = "kf.dev/v1alpha1"
	brokeredInstance.Kind = "ServiceInstance"
	brokeredInstance.Spec.Brokered = &v1alpha1.BrokeredInstance{
		Broker: "brokered-sb",
	}

	osbInstance := &v1alpha1.ServiceInstance{}
	osbInstance.Name = "osb-instance"
	osbInstance.Namespace = "test-ns"
	osbInstance.APIVersion = "kf.dev/v1alpha1"
	osbInstance.Kind = "ServiceInstance"
	osbInstance.Spec.OSB = &v1alpha1.OSBInstance{
		BrokerName: "osb-sb",
	}

	// Create a fake ServiceInstance lister.
	kfClient := kffake.NewSimpleClientset(brokeredInstance, osbInstance)
	informers := externalversions.NewSharedInformerFactory(kfClient, 0)
	serviceInstanceInformer := informers.Kf().V1alpha1().ServiceInstances().Informer()
	serviceInstanceLister := informers.Kf().V1alpha1().ServiceInstances().Lister()

	informers.Start(ctx.Done())
	cache.WaitForCacheSync(ctx.Done(), serviceInstanceInformer.HasSynced)

	brokeredServiceBroker := &v1alpha1.ClusterServiceBroker{}
	brokeredServiceBroker.Name = "brokered-sb"

	OSBServiceBroker := &v1alpha1.ClusterServiceBroker{}
	OSBServiceBroker.Name = "osb-sb"

	deletableServiceBroker := &v1alpha1.ClusterServiceBroker{}
	deletableServiceBroker.Name = "deletable-sb"

	cases := map[string]struct {
		clusterservicebroker *v1alpha1.ClusterServiceBroker
		want                 error
	}{
		"brokered instance exists": {
			clusterservicebroker: brokeredServiceBroker,
			want:                 errors.New("ServiceInstance \"brokered-instance\" at Space \"test-ns\" still exists for broker \"brokered-sb\""),
		},
		"OSB instance exists": {
			clusterservicebroker: OSBServiceBroker,
			want:                 errors.New("ServiceInstance \"osb-instance\" at Space \"test-ns\" still exists for broker \"osb-sb\""),
		},
		"no matching instance exists": {
			clusterservicebroker: deletableServiceBroker,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := validateClusterServiceBroker(serviceInstanceLister, tc.clusterservicebroker)
			testutil.AssertErrorsEqual(t, tc.want, got)
		})
	}
}
