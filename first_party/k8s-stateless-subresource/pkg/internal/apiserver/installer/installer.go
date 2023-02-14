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

package installer

import (
	"errors"
	"fmt"
	"path"
	"time"

	"k8s.io/apiserver/pkg/endpoints"
	"k8s.io/apiserver/pkg/endpoints/discovery"

	"github.com/emicklei/go-restful/v3"
)

// APIGroupVersion mirrors endpoints.APIGroupVersion. However, it also allows
// a WebService to get routes registered. At some point (ideally) this would
// be replaced by k8s.io/apiserver/pkg/endpoints.
type APIGroupVersion struct {
	// ResourceLister lists the API server's resource.
	ResourceLister discovery.APIResourceLister

	// Register is invoked to add the user's WebService to the container.
	Register func(*restful.WebService) error

	*endpoints.APIGroupVersion
}

// InstallREST registers the dynamic REST handlers into a restful Container.
// It is expected that the provided path root prefix will serve all
// operations. Root MUST NOT end in a slash. It should mirror InstallREST in
// the plain APIGroupVersion.
func (g *APIGroupVersion) InstallREST(container *restful.Container) error {
	lister := g.ResourceLister
	if lister == nil {
		return errors.New("must provide a dynamic lister for dynamic API groups")
	}

	installer := g.newDynamicInstaller()
	ws := installer.NewWebService()

	err := installer.install(ws)
	versionDiscoveryHandler := discovery.NewAPIVersionHandler(g.Serializer, g.GroupVersion, lister)
	versionDiscoveryHandler.AddToWebService(ws)
	container.Add(ws)
	return err
}

// newDynamicInstaller is a helper to create the installer. It mirrors
// newInstaller in APIGroupVersion.
func (g *APIGroupVersion) newDynamicInstaller() *APIInstaller {
	prefix := path.Join(g.Root, g.GroupVersion.Group, g.GroupVersion.Version)
	installer := &APIInstaller{
		group:             g,
		prefix:            prefix,
		minRequestTimeout: g.MinRequestTimeout,
		register:          g.Register,
	}

	return installer
}

// APIInstaller is a specialized API installer for the metrics API.  It is
// intended to be fully compliant with the Kubernetes API server conventions,
// but serves wildcard resource/subresource routes instead of hard-coded
// resources and subresources.
type APIInstaller struct {
	group             *APIGroupVersion
	prefix            string
	minRequestTimeout time.Duration
	register          func(*restful.WebService) error
}

// install installs handlers for External Metrics API resources.
func (a *APIInstaller) install(ws *restful.WebService) error {
	if err := a.register(ws); err != nil {
		return fmt.Errorf("error in registering resource: %v", err)
	}

	return nil
}

// NewWebService creates a new restful webservice with the api installer's
// prefix and version.
func (a *APIInstaller) NewWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(a.prefix)
	ws.Doc("API at " + a.prefix)
	ws.Consumes("*/*")
	ws.Produces("*/*")
	ws.ApiVersion(a.group.GroupVersion.String())

	return ws
}
