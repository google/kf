// Copyright 2026 Google LLC
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

// This file was generated with fakelister/generator.go, DO NOT EDIT IT.

package apiservercerts

import (
	objectpackage "k8s.io/api/core/v1"
	listerpackage "k8s.io/client-go/listers/core/v1"

	"context"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"reflect"
	"sort"
	"sync"
	"testing"
)

type SecretKey struct{}

func withFakeSecretLister(ctx context.Context) context.Context {
	return context.WithValue(ctx, SecretKey{}, &fakeSecretLister{})
}

func fakeSecretListerFromContext(ctx context.Context) *fakeSecretLister {
	return ctx.Value(SecretKey{}).(*fakeSecretLister)
}

// fakeSecretLister implements SecretLister. We can't use the normal K8s fakes
// because the listers use caches. Therefore they don't interact with the
// FakeClientSet quite right (atleast as best as I can tell).
type fakeSecretLister struct {
	mu sync.RWMutex
	listerpackage.SecretLister
	listerpackage.SecretNamespaceLister
	namespace string

	items map[string]map[string]*objectpackage.Secret
	err   error
}

// Add adds the object into the fake. The object's name and namespace are used
// to determine when the object should be returned.
func (f *fakeSecretLister) Add(x *objectpackage.Secret) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.items == nil {
		f.items = make(map[string]map[string]*objectpackage.Secret)
	}

	m := f.items[x.Namespace]
	if m == nil {
		m = make(map[string]*objectpackage.Secret)
		f.items[x.Namespace] = m
	}
	m[x.Name] = x
}

// AssertCacheIsPreserved should be invoked before the test is started. It
// ensures that the objects in the cache are never altered.
func (f *fakeSecretLister) AssertCacheIsPreserved(t *testing.T) {
	orig := f.flatten()
	t.Cleanup(func() {
		if !reflect.DeepEqual(orig, f.flatten()) {
			t.Fatal("cached objects must not be altered")
		}
	})
}

// flatten will return deepCopies of each object. This is useful for asserting
// that the objects weren't changed. Cached objects should never be updated.
// The resulting slice is sorted and deterministic.
func (f *fakeSecretLister) flatten() []*objectpackage.Secret {
	f.mu.RLock()
	var items []*objectpackage.Secret
	for ns := range f.items {
		for _, x := range f.items[ns] {
			items = append(items, x.DeepCopy())
		}
	}
	f.mu.RUnlock()

	// Sort the objects by namespace and name to ensure each list is
	// deterministic.
	sort.Slice(items, func(i, j int) bool {
		a, b := items[i], items[j]
		keyA, keyB := a.Namespace+"/"+a.Name, b.Namespace+"/"+b.Name
		return keyA < keyB
	})

	return items
}

func (f *fakeSecretLister) Secrets(ns string) listerpackage.SecretNamespaceLister {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.namespace = ns
	return f
}

// Get implements the Lister interface.
func (f *fakeSecretLister) Get(name string) (ret *objectpackage.Secret, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.err != nil {
		return nil, f.err
	}

	if x, ok := f.items[f.namespace][name]; ok {
		return x, nil
	}

	return nil, apierrs.NewNotFound(objectpackage.Resource("Secret"), name)
}

// List implements the Lister interface.
func (f *fakeSecretLister) List(selector labels.Selector) (ret []*objectpackage.Secret, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, t := range f.items[f.namespace] {
		ret = append(ret, t)
	}

	return ret, f.err
}
