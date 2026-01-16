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

package main

import (
	"io"

	"context"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/k8s-stateless-subresource/pkg/apiserver"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	versioned "github.com/google/kf/v2/pkg/client/kf/clientset/versioned"
	kfinformer "github.com/google/kf/v2/pkg/client/kf/informers/externalversions/kf/v1alpha1"
	kfclient "github.com/google/kf/v2/pkg/client/kf/injection/client"
	buildinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/build"
	sourcepackageinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/sourcepackage"
	spaceinformer "github.com/google/kf/v2/pkg/client/kf/injection/informers/kf/v1alpha1/space"
	"github.com/google/kf/v2/pkg/sourceimage"
	flag "github.com/spf13/pflag"
	"github.com/tektoncd/pipeline/pkg/apis/config"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/util/retry"
	"knative.dev/pkg/logging"
)

func main() {
	// Setup the flags used to configure the API server.
	cfg := apiserver.NewConfig()
	cfg.AddFlags(flag.CommandLine)
	cfg.LongRunningFunc = func(r *http.Request, requestInfo *request.RequestInfo) bool {
		// All requests should be considered long running.
		return true
	}
	flag.Parse()

	apiserver.MainWithConfig(
		context.Background(),
		"upload-api-server",
		cfg,
		apiserver.EndpointConstructorStruct{
			Scheme: scheme,
			Resource: metav1.APIResource{
				Name:       schema.GroupResource{Group: groupName, Resource: "sourcepackages"}.String(),
				Group:      groupName,
				Namespaced: true,
				Kind:       "Upload",
				Verbs: metav1.Verbs{
					"get",
					"create",
				},
			},
			RegisterWebServiceF: func(ctx context.Context, ws *restful.WebService) error {
				logger := logging.FromContext(ctx)
				buildInformer := buildinformer.Get(ctx)
				sourcePackageInformer := sourcepackageinformer.Get(ctx)
				spaceInformer := spaceinformer.Get(ctx)
				kfClient := kfclient.Get(ctx)

				handler := &handler{
					logger:                logger,
					cfg:                   config.FromContextOrDefaults(ctx).Defaults,
					buildInformer:         buildInformer,
					spaceInformer:         spaceInformer,
					sourcePackageInformer: sourcePackageInformer,
					kfClient:              kfClient,
				}

				{
					path := "proxy/namespaces/{namespace}/{subresource}"
					route := ws.POST(path).To(handler.post).
						Doc("Upload data for a SourcePackage").
						Operation("upload").
						Consumes("application/octet-stream").
						Produces("application/json")
					ws.Route(route)
				}

				{
					path := "proxy/namespaces/{namespace}/{subresource}"
					route := ws.GET(path).To(handler.get).
						Doc("Get data for a SourcePackage").
						Operation("get").
						Produces("application/octet-stream")
					ws.Route(route)
				}

				{
					path := "proxy/namespaces/{namespace}/{subresource}/image"
					route := ws.POST(path).To(handler.uploadImage).
						Doc("Upload result container for a SourcePackage").
						Operation("upload").
						Consumes("application/octet-stream").
						Produces("application/json")
					ws.Route(route)
				}

				return nil
			},
		},
	)
}

type handler struct {
	cfg                   *config.Defaults
	logger                *zap.SugaredLogger
	buildInformer         kfinformer.BuildInformer
	spaceInformer         kfinformer.SpaceInformer
	sourcePackageInformer kfinformer.SourcePackageInformer
	kfClient              versioned.Interface
}

func (h *handler) post(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	ns := req.PathParameter("namespace")
	sr := req.PathParameter("subresource")

	// Always close the body, but don't worry about the error.
	defer req.Request.Body.Close()

	u := sourceimage.NewUploader(
		h.spaceInformer.Lister(),
		h.sourcePackageInformer.Lister(),
		func(path, imageName string) (name.Reference, error) {
			img, err := sourceimage.PackageFile(path)
			if err != nil {
				return nil, err
			}

			return sourceimage.PushImage(imageName, img, true)
		},

		func(s *v1alpha1.SourcePackage) error {
			return retry.OnError(retry.DefaultBackoff, func(e error) bool {
				// always retry errors until timed out
				return true
			}, func() error {
				if _, err := h.kfClient.
					KfV1alpha1().
					SourcePackages(ns).
					UpdateStatus(ctx, s, metav1.UpdateOptions{}); err != nil {
					return err
				}

				return nil

			})
		},
	)
	// Retrying 13 times should take around 2.73 minutes.
	const maxRetriesForGetSourcePackage = 13
	if _, err := u.Upload(ctx, ns, sr, maxRetriesForGetSourcePackage, req.Request.Body); err != nil {
		h.logger.Warnf("failed to save image %s/%s: %v", ns, sr, err)
		resp.WriteErrorString(http.StatusInternalServerError, "failed to save image")
		return
	}
}

func (h *handler) get(req *restful.Request, resp *restful.Response) {
	ns := req.PathParameter("namespace")
	sr := req.PathParameter("subresource")

	defer req.Request.Body.Close()

	if extract, err := sourceimage.Download(h.sourcePackageInformer.Lister(), ns, sr); err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
	} else {
		defer extract.Close()
		if _, err := io.Copy(resp, extract); err != nil {
			resp.WriteError(http.StatusInternalServerError, err)
		}
	}
}

func (h *handler) uploadImage(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	ns := req.PathParameter("namespace")
	sr := req.PathParameter("subresource")

	// Always close the body, but don't worry about the error.
	defer req.Request.Body.Close()

	u := sourceimage.NewImageUploader(
		h.cfg,
		h.buildInformer.Lister(),
		h.spaceInformer.Lister(),
		func(path, imageName string) (name.Reference, error) {
			tag, err := name.NewTag(imageName, name.WeakValidation)
			if err != nil {
				return nil, err
			}

			img, err := tarball.ImageFromPath(path, &tag)
			if err != nil {
				return nil, err
			}

			return sourceimage.PushImage(imageName, img, true)
		},
	)
	imageName, err := u.Upload(ctx, ns, sr, req.Request.Body)
	if err != nil {
		h.logger.Warnf("failed to save image %s/%s: %v", ns, sr, err)
		resp.WriteErrorString(http.StatusInternalServerError, "failed to save image")
		return
	}
	resp.WriteAsJson(map[string]string{
		"digest": imageName.Name(),
	})
}
