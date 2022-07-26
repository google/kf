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

package apiserver_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/google/k8s-stateless-subresource/pkg/internal/apiserver"
	nodev1 "k8s.io/api/node/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/server"
)

func TestServer_InstallAPI(t *testing.T) {
	t.Parallel()
	container := &restful.Container{
		ServeMux: http.NewServeMux(),
	}
	s := apiserver.Server{
		GenericAPIServer: &server.GenericAPIServer{
			Handler: &server.APIServerHandler{
				GoRestfulContainer: container,
			},
		},
	}

	// Use node schema for test.
	scheme := runtime.NewScheme()
	utilruntime.Must(nodev1.AddToScheme(scheme))

	codecs := serializer.NewCodecFactory(scheme)
	resource := metav1.APIResource{}
	register := func(*restful.WebService) error { return nil }

	s.InstallAPI(scheme, codecs, resource, register)

	// Assert that the route was setup on the container.
	if actual, expected := len(container.RegisteredWebServices()), 1; actual != expected {
		t.Fatalf("expected %d, got %d", expected, actual)
	}

	ws := container.RegisteredWebServices()[0]

	if actual, expected := len(ws.Routes()), 1; actual != expected {
		t.Fatalf("expected %d, got %d", expected, actual)
	}

	r := ws.Routes()[0]

	if actual, expected := r.Method, "GET"; actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
	if actual, expected := r.Path, "/apis/v1/"; actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
	if actual, expected := r.Doc, "get available resources"; actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
	if actual, expected := r.Operation, "getAPIResources"; actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
	if actual, expected := r.Produces, []string{"application/json", "application/yaml", "application/vnd.kubernetes.protobuf"}; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
	if actual, expected := r.Consumes, []string{"application/json", "application/yaml", "application/vnd.kubernetes.protobuf"}; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
