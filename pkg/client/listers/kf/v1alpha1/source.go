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
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// SourceLister helps list Sources.
type SourceLister interface {
	// List lists all Sources in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Source, err error)
	// Sources returns an object that can list and get Sources.
	Sources(namespace string) SourceNamespaceLister
	SourceListerExpansion
}

// sourceLister implements the SourceLister interface.
type sourceLister struct {
	indexer cache.Indexer
}

// NewSourceLister returns a new SourceLister.
func NewSourceLister(indexer cache.Indexer) SourceLister {
	return &sourceLister{indexer: indexer}
}

// List lists all Sources in the indexer.
func (s *sourceLister) List(selector labels.Selector) (ret []*v1alpha1.Source, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Source))
	})
	return ret, err
}

// Sources returns an object that can list and get Sources.
func (s *sourceLister) Sources(namespace string) SourceNamespaceLister {
	return sourceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SourceNamespaceLister helps list and get Sources.
type SourceNamespaceLister interface {
	// List lists all Sources in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Source, err error)
	// Get retrieves the Source from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Source, error)
	SourceNamespaceListerExpansion
}

// sourceNamespaceLister implements the SourceNamespaceLister
// interface.
type sourceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Sources in the indexer for a given namespace.
func (s sourceNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Source, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Source))
	})
	return ret, err
}

// Get retrieves the Source from the indexer for a given namespace and name.
func (s sourceNamespaceLister) Get(name string) (*v1alpha1.Source, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("source"), name)
	}
	return obj.(*v1alpha1.Source), nil
}
