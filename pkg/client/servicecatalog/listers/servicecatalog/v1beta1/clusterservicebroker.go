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

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ClusterServiceBrokerLister helps list ClusterServiceBrokers.
type ClusterServiceBrokerLister interface {
	// List lists all ClusterServiceBrokers in the indexer.
	List(selector labels.Selector) (ret []*v1beta1.ClusterServiceBroker, err error)
	// Get retrieves the ClusterServiceBroker from the index for a given name.
	Get(name string) (*v1beta1.ClusterServiceBroker, error)
	ClusterServiceBrokerListerExpansion
}

// clusterServiceBrokerLister implements the ClusterServiceBrokerLister interface.
type clusterServiceBrokerLister struct {
	indexer cache.Indexer
}

// NewClusterServiceBrokerLister returns a new ClusterServiceBrokerLister.
func NewClusterServiceBrokerLister(indexer cache.Indexer) ClusterServiceBrokerLister {
	return &clusterServiceBrokerLister{indexer: indexer}
}

// List lists all ClusterServiceBrokers in the indexer.
func (s *clusterServiceBrokerLister) List(selector labels.Selector) (ret []*v1beta1.ClusterServiceBroker, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.ClusterServiceBroker))
	})
	return ret, err
}

// Get retrieves the ClusterServiceBroker from the index for a given name.
func (s *clusterServiceBrokerLister) Get(name string) (*v1beta1.ClusterServiceBroker, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("clusterservicebroker"), name)
	}
	return obj.(*v1beta1.ClusterServiceBroker), nil
}
