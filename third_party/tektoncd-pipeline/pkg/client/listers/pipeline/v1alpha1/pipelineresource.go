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

package v1alpha1

import (
	v1alpha1 "github.com/google/kf/third_party/tektoncd-pipeline/pkg/apis/pipeline/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// PipelineResourceLister helps list PipelineResources.
type PipelineResourceLister interface {
	// List lists all PipelineResources in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.PipelineResource, err error)
	// PipelineResources returns an object that can list and get PipelineResources.
	PipelineResources(namespace string) PipelineResourceNamespaceLister
	PipelineResourceListerExpansion
}

// pipelineResourceLister implements the PipelineResourceLister interface.
type pipelineResourceLister struct {
	indexer cache.Indexer
}

// NewPipelineResourceLister returns a new PipelineResourceLister.
func NewPipelineResourceLister(indexer cache.Indexer) PipelineResourceLister {
	return &pipelineResourceLister{indexer: indexer}
}

// List lists all PipelineResources in the indexer.
func (s *pipelineResourceLister) List(selector labels.Selector) (ret []*v1alpha1.PipelineResource, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.PipelineResource))
	})
	return ret, err
}

// PipelineResources returns an object that can list and get PipelineResources.
func (s *pipelineResourceLister) PipelineResources(namespace string) PipelineResourceNamespaceLister {
	return pipelineResourceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PipelineResourceNamespaceLister helps list and get PipelineResources.
type PipelineResourceNamespaceLister interface {
	// List lists all PipelineResources in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.PipelineResource, err error)
	// Get retrieves the PipelineResource from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.PipelineResource, error)
	PipelineResourceNamespaceListerExpansion
}

// pipelineResourceNamespaceLister implements the PipelineResourceNamespaceLister
// interface.
type pipelineResourceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all PipelineResources in the indexer for a given namespace.
func (s pipelineResourceNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.PipelineResource, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.PipelineResource))
	})
	return ret, err
}

// Get retrieves the PipelineResource from the indexer for a given namespace and name.
func (s pipelineResourceNamespaceLister) Get(name string) (*v1alpha1.PipelineResource, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("pipelineresource"), name)
	}
	return obj.(*v1alpha1.PipelineResource), nil
}