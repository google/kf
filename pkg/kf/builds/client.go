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

package builds

import (
	"fmt"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/doctor"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClientInterface is the main interface for interacting with Knative builds.
//
// It's built to be generic enough that we could swap in alternative
// implementations like Tekton without changing too much.
type ClientInterface interface {
	doctor.Diagnosable

	Create(name string, template build.TemplateInstantiationSpec, opts ...CreateOption) (*build.Build, error)
	Status(name string, opts ...StatusOption) (complete bool, err error)
	Delete(name string, opts ...DeleteOption) error

	// TODO(josephlewis42): log tailer should be brought into this package
}

var _ ClientInterface = (*Client)(nil)

// NewClient creates a new build client.
func NewClient(buildClient cbuild.BuildV1alpha1Interface) ClientInterface {
	return &Client{
		buildClient: buildClient,
	}
}

// Client is a client to knative.Build built in a way that other systems could
// be mostly dropped in as replacements.
type Client struct {
	buildClient cbuild.BuildV1alpha1Interface
}

// Create creates a new build.
func (c *Client) Create(name string, template build.TemplateInstantiationSpec, opts ...CreateOption) (*build.Build, error) {
	build := PopulateTemplate(name, template, opts...)

	return c.buildClient.Builds(build.Namespace).Create(build)
}

// Status gets the status of the build with the given name by calling BuildStatus.
// If the bulid doesn't exist an error is returned.
func (c *Client) Status(name string, opts ...StatusOption) (bool, error) {
	cfg := StatusOptionDefaults().Extend(opts).toConfig()

	bld, err := c.buildClient.Builds(cfg.Namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return true, fmt.Errorf("couldn't get build %q, %s", name, err.Error())
	}

	return BuildStatus(*bld)
}

// Delete removes a build.
func (c *Client) Delete(name string, opts ...DeleteOption) error {
	cfg := DeleteOptionDefaults().Extend(opts).toConfig()

	return c.buildClient.Builds(cfg.Namespace).Delete(name, nil)
}

func (c *Client) Diagnose(d *doctor.Diagnostic) {
	d.Run("ClusterBuildTemplates", func(d *doctor.Diagnostic) {
		for _, template := range clusterBuiltins() {
			d.Run(template.Name, func(d *doctor.Diagnostic) {
				_, err := c.buildClient.ClusterBuildTemplates().Get(template.Name, v1.GetOptions{})
				if err != nil {
					d.Fatalf("Error fetching template: %v", err)
				}
			})
		}
	})
}
