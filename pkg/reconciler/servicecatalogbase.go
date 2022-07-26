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

package reconciler

import (
	"context"
	"errors"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kfclusterservicebrokerinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/clusterservicebroker"
	kfservicebrokerinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/servicebroker"
	kfserviceinstanceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/serviceinstance"
	kfserviceinstancebindinginformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/serviceinstancebinding"
	kflisters "github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/osbutil"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/configmap"
	osbclient "sigs.k8s.io/go-open-service-broker-client/v2"
)

// ServiceCatalogBase contains utilities used by service brokers.
type ServiceCatalogBase struct {
	*Base

	KfClusterServiceBrokerLister   kflisters.ClusterServiceBrokerLister
	KfServiceBrokerLister          kflisters.ServiceBrokerLister
	KfServiceInstanceLister        kflisters.ServiceInstanceLister
	KfServiceInstanceBindingLister kflisters.ServiceInstanceBindingLister
}

// NewServiceCatalogBase instantiates a new instance of ServiceCatalogBase.
func NewServiceCatalogBase(ctx context.Context, cmw configmap.Watcher) *ServiceCatalogBase {
	clusterServiceBrokerInformer := kfclusterservicebrokerinformer.Get(ctx)
	serviceBrokerInformer := kfservicebrokerinformer.Get(ctx)
	serviceInstanceInformer := kfserviceinstanceinformer.Get(ctx)
	serviceInstanceBindingInformer := kfserviceinstancebindinginformer.Get(ctx)

	return &ServiceCatalogBase{
		Base: NewBase(ctx, cmw),

		KfClusterServiceBrokerLister:   clusterServiceBrokerInformer.Lister(),
		KfServiceBrokerLister:          serviceBrokerInformer.Lister(),
		KfServiceInstanceLister:        serviceInstanceInformer.Lister(),
		KfServiceInstanceBindingLister: serviceInstanceBindingInformer.Lister(),
	}
}

// GetInstanceForBinding returns a ServiceInstance that the binding belongs to.
func (scb *ServiceCatalogBase) GetInstanceForBinding(
	binding *v1alpha1.ServiceInstanceBinding,
) (*v1alpha1.ServiceInstance, error) {
	instanceName := binding.Spec.InstanceRef.Name
	return scb.KfServiceInstanceLister.
		ServiceInstances(binding.Namespace).
		Get(instanceName)
}

// GetBrokerForInstance returns a ServiceBroker that the ServiceInstance belongs
// to.
func (scb *ServiceCatalogBase) GetBrokerForInstance(
	instance *v1alpha1.ServiceInstance,
) (broker v1alpha1.CommonServiceBroker, err error) {
	if !instance.IsKfBrokered() {
		return nil, errors.New("service isn't a backed by a service broker")
	}

	brokerName := instance.Spec.OSB.BrokerName

	// check the broker type, return appropriate one
	if instance.Spec.OSB.Namespaced {
		return scb.KfServiceBrokerLister.
			ServiceBrokers(instance.Namespace).
			Get(brokerName)
	}

	return scb.KfClusterServiceBrokerLister.Get(brokerName)
}

// GetCredentialsSecretForBroker returns the credentials secret for the given
// broker.
func (scb *ServiceCatalogBase) GetCredentialsSecretForBroker(
	broker v1alpha1.CommonServiceBroker,
) (*corev1.Secret, error) {
	credsRef := broker.GetCredentialsSecretRef()
	return scb.SecretLister.Secrets(credsRef.Namespace).Get(credsRef.Name)
}

// GetClientForServiceInstance gets the OSB client for a specific ServiceInstance.
func (scb *ServiceCatalogBase) GetClientForServiceInstance(serviceInstance *v1alpha1.ServiceInstance) (osbclient.Client, error) {
	broker, err := scb.GetBrokerForInstance(serviceInstance)
	if err != nil {
		return nil, err
	}

	brokerCreds, err := scb.GetCredentialsSecretForBroker(broker)
	if err != nil {
		return nil, err
	}

	return osbutil.NewClient(brokerCreds)
}
