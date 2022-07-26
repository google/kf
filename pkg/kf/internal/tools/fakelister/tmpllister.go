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

package fakelister

import (
	"text/template"

	"github.com/google/kf/v2/pkg/kf/internal/tools/generator"
)

var listerTemplate = template.Must(template.New("").Funcs(generator.TemplateFuncs()).Parse(`
{{genlicense}}

{{gennotice "fakelister/generator.go"}}

package {{.Package}}

import (
  objectpackage "{{.ObjectPackage}}"
  listerpackage "{{.ListerPackage}}"

  "reflect"
  "testing"
  "context"
  "sort"
  "sync"
  "k8s.io/apimachinery/pkg/labels"
  apierrs "k8s.io/apimachinery/pkg/api/errors"
)

type {{.ObjectType}}Key struct{}

func withFake{{.ObjectType}}Lister(ctx context.Context) context.Context {
	return context.WithValue(ctx, {{.ObjectType}}Key{}, &fake{{.ObjectType}}Lister{})
}

func fake{{.ObjectType}}ListerFromContext(ctx context.Context) *fake{{.ObjectType}}Lister{
	return ctx.Value({{.ObjectType}}Key{}).(*fake{{.ObjectType}}Lister)
}

// fake{{.ObjectType}}Lister implements {{.ObjectType}}Lister. We can't use the normal K8s fakes
// because the listers use caches. Therefore they don't interact with the
// FakeClientSet quite right (atleast as best as I can tell).
type fake{{.ObjectType}}Lister struct {
    mu sync.RWMutex
	listerpackage.{{.ObjectType}}Lister
	{{if .Namespaced}}listerpackage.{{.ObjectType}}NamespaceLister
	{{else}}
	// namespace is never set and therefore always empty. It is here so the
	// generator can use the same logic as it would with namespaced objects.
    {{end}}namespace string

	items map[string]map[string]*objectpackage.{{.ObjectType}}
	err    error
}

// Add adds the object into the fake. The object's name and namespace are used
// to determine when the object should be returned.
func (f *fake{{.ObjectType}}Lister) Add(x *objectpackage.{{.ObjectType}}) {
    f.mu.Lock()
	defer f.mu.Unlock()
	if f.items == nil {
		f.items = make(map[string]map[string]*objectpackage.{{.ObjectType}})
	}

	m := f.items[x.Namespace]
	if m == nil {
		m = make(map[string]*objectpackage.{{.ObjectType}})
		f.items[x.Namespace] = m
	}
	m[x.Name] = x
}

// AssertCacheIsPreserved should be invoked before the test is started. It
// ensures that the objects in the cache are never altered.
func (f *fake{{.ObjectType}}Lister) AssertCacheIsPreserved(t *testing.T) {
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
func (f *fake{{.ObjectType}}Lister) flatten() []*objectpackage.{{.ObjectType}} {
    f.mu.RLock()
	var items []*objectpackage.{{.ObjectType}}
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

{{if .Namespaced}}
func (f *fake{{.ObjectType}}Lister) {{.ObjectType}}s(ns string) listerpackage.{{.ObjectType}}NamespaceLister {
    f.mu.Lock()
	defer f.mu.Unlock()
	f.namespace = ns
	return f
}
{{end}}

// Get implements the Lister interface.
func (f *fake{{.ObjectType}}Lister) Get(name string) (ret *objectpackage.{{.ObjectType}}, err error) {
    f.mu.RLock()
	defer f.mu.RUnlock()
    if f.err != nil {
		return nil, f.err
	}

	if x, ok := f.items[f.namespace][name]; ok {
		return x, nil
	}

	return nil, apierrs.NewNotFound(objectpackage.Resource("{{.ObjectType}}"), name)
}

// List implements the Lister interface.
func (f *fake{{.ObjectType}}Lister) List(selector labels.Selector) (ret []*objectpackage.{{.ObjectType}}, err error) {
    f.mu.RLock()
	defer f.mu.RUnlock()
	for _, t := range f.items[f.namespace] {
		ret = append(ret, t)
	}

	return ret, f.err
}
`))
