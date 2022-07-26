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

package tableclient

import (
	"context"
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// Type contains type information for the table client.
type Type interface {
	// Namespaced returns whether the resource is namespace-scoped.
	Namespaced() bool
	// GroupVersionResource returns the GVR of the target resource.
	GroupVersionResource(ctx context.Context) schema.GroupVersionResource
}

// Interface contains the public definitions for table clients.
type Interface interface {
	Table(ctx context.Context, typ Type, namespace string, opts metav1.ListOptions) (*metav1beta1.Table, error)
}

// New creates a table client given a RestConfig. The config WILL BE MODIFIED.
func New(config *rest.Config) (Interface, error) {
	config.Wrap(NewTableRoundTripper)

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &tableClientImpl{
		client: client,
	}, nil
}

type tableClientImpl struct {
	client dynamic.Interface
}

// Table gets a table formatted version of the given type. Namespace is
// ignored if the type's Namespaced() function returns false.
func (t *tableClientImpl) Table(ctx context.Context, typ Type, namespace string, opts metav1.ListOptions) (*metav1beta1.Table, error) {
	var resourceFetcher dynamic.ResourceInterface
	if typ.Namespaced() {
		resourceFetcher = t.client.Resource(typ.GroupVersionResource(ctx)).Namespace(namespace)
	} else {
		resourceFetcher = t.client.Resource(typ.GroupVersionResource(ctx))
	}

	unstr, err := resourceFetcher.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	table := &metav1beta1.Table{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstr.Object, table); err != nil {
		return nil, err
	}

	return table, nil
}

// NewTableRoundTripper creates a RoundTripper that requests resource lists
// as tables from the Kubernetes API server.
// See https://kubernetes.io/docs/reference/using-api/api-concepts/#receiving-resources-as-tables
// for more information.
func NewTableRoundTripper(parent http.RoundTripper) http.RoundTripper {
	return &tableRoundTripper{
		parent: parent,
	}
}

type tableRoundTripper struct {
	parent http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (trt *tableRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	group := metav1beta1.GroupName
	version := metav1beta1.SchemeGroupVersion.Version

	tableParam := fmt.Sprintf("application/json;as=Table;v=%s;g=%s, application/json", version, group)
	req.Header.Set("Accept", tableParam)
	return trt.parent.RoundTrip(req)
}
