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

package installer_test

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/google/k8s-stateless-subresource/pkg/internal/apiserver/installer"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/endpoints"
	"k8s.io/apiserver/pkg/endpoints/discovery"
	"k8s.io/apiserver/pkg/server"
)

func TestAPIGroupVersion_InstallREST(t *testing.T) {
	t.Parallel()

	container := &restful.Container{
		ServeMux: http.NewServeMux(),
	}
	gv := apiGV(nil)

	err := gv.InstallREST(container)

	if err != nil {
		t.Fatal(err)
	}

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
	if actual, expected := r.Path, "/apis/"; actual != expected {
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

func TestAPIGroupVersion_InstallREST_registerErr(t *testing.T) {
	t.Parallel()

	container := &restful.Container{
		ServeMux: http.NewServeMux(),
	}
	gv := apiGV(errors.New("some-error"))

	err := gv.InstallREST(container)

	if actual, expected := fmt.Sprintf("%v", err), "error in registering resource: some-error"; actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

func TestAPIGroupVersion_InstallREST_errorForEmptyResourceLister(t *testing.T) {
	t.Parallel()

	gv := apiGV(nil)
	gv.ResourceLister = nil

	err := gv.InstallREST(nil)

	if actual, expected := fmt.Sprintf("%v", err), "must provide a dynamic lister for dynamic API groups"; actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

func apiGV(registerErr error) *installer.APIGroupVersion {
	resource := metav1.APIResource{}
	groupVersion := schema.GroupVersion{}
	register := func(*restful.WebService) error {
		return registerErr
	}

	scheme := runtime.NewScheme()
	codec := serializer.NewCodecFactory(scheme)

	groupInfo := server.NewDefaultAPIGroupInfo(
		resource.Group,
		scheme,
		runtime.NewParameterCodec(scheme),
		codec,
	)

	return &installer.APIGroupVersion{
		APIGroupVersion: &endpoints.APIGroupVersion{
			Root:             server.APIGroupPrefix,
			GroupVersion:     groupVersion,
			MetaGroupVersion: groupInfo.MetaGroupVersion,

			ParameterCodec:  groupInfo.ParameterCodec,
			Serializer:      groupInfo.NegotiatedSerializer,
			Creater:         groupInfo.Scheme,
			Convertor:       groupInfo.Scheme,
			UnsafeConvertor: runtime.UnsafeObjectConvertor(groupInfo.Scheme),
			Typer:           groupInfo.Scheme,
			Linker:          runtime.SelfLinker(meta.NewAccessor()),
		},

		ResourceLister: discovery.APIResourceListerFunc(func() []metav1.APIResource {
			return []metav1.APIResource{resource}
		}),
		Register: register,
	}
}
