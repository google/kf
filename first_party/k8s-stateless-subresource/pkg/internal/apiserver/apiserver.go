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

package apiserver

import (
	"github.com/emicklei/go-restful"
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

// Server contains state for a Kubernetes cluster master/api server.
type Server struct {
	GenericAPIServer *server.GenericAPIServer
}

// InstallAPI registers each API version for the given resource group.
func (s *Server) InstallAPI(
	scheme *runtime.Scheme,
	codecs serializer.CodecFactory,
	resource metav1.APIResource,
	register func(*restful.WebService) error,
) error {
	groupInfo := server.NewDefaultAPIGroupInfo(
		resource.Group,
		scheme,
		runtime.NewParameterCodec(scheme),
		codecs,
	)
	container := s.GenericAPIServer.Handler.GoRestfulContainer

	// Register custom metrics REST handler for all supported API versions.
	for _, mainGroupVer := range groupInfo.PrioritizedVersions {
		apiGV := s.apiGV(resource, &groupInfo, mainGroupVer, register)
		if err := apiGV.InstallREST(container); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) apiGV(
	resource metav1.APIResource,
	groupInfo *server.APIGroupInfo,
	groupVersion schema.GroupVersion,
	register func(*restful.WebService) error,
) *installer.APIGroupVersion {
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
