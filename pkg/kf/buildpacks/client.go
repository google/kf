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

package buildpacks

import (
	"encoding/json"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	gcrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/kf/pkg/kf/doctor"
)

// Client is the main interface for interacting with Buildpacks.
type Client interface {
	doctor.Diagnosable

	// List lists the buildpacks available on the given builder image.
	List(builderImage string) ([]Buildpack, error)

	// Stacks lists the stacks available on the given builder image.
	Stacks(builderImage string) ([]string, error)
}

// Buildpack has the information from a Buildpack Builder.
type Buildpack struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Latest  bool   `json:"latest"`
}

// RemoteImageFetcher is implemented by
// github.com/google/go-containerregistry/pkg/v1/remote.Image
type RemoteImageFetcher func(ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error)

type client struct {
	imageFetcher RemoteImageFetcher
}

// NewClient creates a new Client.
func NewClient(
	imageFetcher RemoteImageFetcher,
) Client {
	return &client{
		imageFetcher: imageFetcher,
	}
}

const metadataLabel = "io.buildpacks.builder.metadata"

// List lists the available buildpacks.
func (c *client) List(builderImage string) ([]Buildpack, error) {
	cfg, err := c.fetchConfig(builderImage)
	if err != nil || cfg == nil {
		return nil, err
	}

	var order struct {
		Buildpacks []Buildpack `json:"buildpacks"`
	}
	if err := json.NewDecoder(strings.NewReader(cfg.Config.Labels[metadataLabel])).Decode(&order); err != nil {
		return nil, err
	}

	return order.Buildpacks, nil
}

// Stacks lists the stacks available.
func (c *client) Stacks(builderImage string) ([]string, error) {
	cfg, err := c.fetchConfig(builderImage)
	if err != nil || cfg == nil {
		return nil, err
	}

	var stack struct {
		Stack struct {
			RunImage struct {
				Image string `json:"image"`
			} `json:"runImage"`
		} `json:"stack"`
	}

	if err := json.NewDecoder(strings.NewReader(cfg.Config.Labels[metadataLabel])).Decode(&stack); err != nil {
		return nil, err
	}

	return []string{stack.Stack.RunImage.Image}, nil
}

func (c *client) fetchConfig(builderImage string) (*gcrv1.ConfigFile, error) {
	imageRef, err := name.ParseReference(builderImage, name.WeakValidation)
	if err != nil {
		return nil, err
	}

	image, err := c.imageFetcher(imageRef, remote.WithAuth(authn.Anonymous))
	if err != nil {
		return nil, err
	}

	cfg, err := image.ConfigFile()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Diagnose checks to validate the cluster's buildpacks.
func (c *client) Diagnose(d *doctor.Diagnostic) {
}
