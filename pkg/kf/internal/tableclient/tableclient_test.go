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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type MockType struct{}

func (MockType) Namespaced() bool {
	return true
}

func (MockType) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "test.group",
		Version:  "v1",
		Resource: "tests",
	}
}

func ExampleNew() {
	client, err := New(&rest.Config{
		Transport: &MockRoundTripper{},
	})
	if err != nil {
		panic(err)
	}

	// This example won't hit a real endpoint, but the mock will show that the URL
	// and selector work for getting namespaced tables.
	if _, err := client.Table(MockType{}, "demo", metav1.ListOptions{}); err != nil {
		panic(err)
	}

	// Output: URL: http://localhost/apis/test.group/v1/namespaces/demo/tests
	// Accepts: application/json;as=Table;v=v1beta1;g=meta.k8s.io, application/json
}

type MockRoundTripper struct{}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Println("URL:", req.URL.String())
	fmt.Println("Accepts:", req.Header.Get("Accept"))

	// Return a mock response
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       ioutil.NopCloser(&bytes.Buffer{}),
	}, nil
}

func ExampleNewTableRoundTripper() {
	rt := NewTableRoundTripper(&MockRoundTripper{})
	req, err := http.NewRequest(http.MethodGet, "master.svc.cluster.local", nil)
	if err != nil {
		panic(err)
	}
	if _, err := rt.RoundTrip(req); err != nil {
		panic(err)
	}

	// Output: URL: master.svc.cluster.local
	// Accepts: application/json;as=Table;v=v1beta1;g=meta.k8s.io, application/json
}
