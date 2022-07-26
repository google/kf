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
	"context"
	"errors"
	"fmt"
	"io"

	cv1alpha1 "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/typed/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
)

// ClientExtension holds additional functions that should be exposed by client.
type ClientExtension interface {
	Tail(ctx context.Context, namespace, name string, writer io.Writer) error
	Status(ctx context.Context, namespace, name string) (bool, error)
}

// BuildTailer tails the Build logs.
type BuildTailer interface {
	Tail(ctx context.Context, out io.Writer, buildName, namespace string) error
}

type buildsClient struct {
	coreClient

	buildTailer BuildTailer
	p           *config.KfParams
}

// NewClient creates a new build client.
func NewClient(p *config.KfParams, kclient cv1alpha1.BuildsGetter, buildTailer BuildTailer) Client {
	return &buildsClient{
		coreClient: coreClient{
			kclient: kclient,
		},
		buildTailer: buildTailer,
		p:           p,
	}
}

// Status gets the status of the build with the given name by calling BuildStatus.
// If the build doesn't exist an error is returned.
func (c *buildsClient) Status(ctx context.Context, namespace, name string) (bool, error) {
	bld, err := c.coreClient.Get(ctx, namespace, name)
	if err != nil {
		return true, err
	}

	return BuildStatus(*bld)
}

// Tail streams the build logs to a local writer.
func (c *buildsClient) Tail(ctx context.Context, namespace, name string, writer io.Writer) error {
	var buildName string
	if c.p.FeatureFlags(ctx).AppDevExperienceBuilds().IsEnabled() {
		// XXX: AppDevExperience doesn't expose Tekton directly, so we'll have
		// to assume that the TaskRun is the same name as the Build.
		buildName = name
	} else {
		bld, err := c.coreClient.Get(ctx, namespace, name)
		if err != nil {
			return err
		}

		buildName = bld.Status.BuildName
		if buildName == "" {
			return errors.New("The build hasn't started yet")
		}
	}

	fmt.Fprintf(writer, "Logs for %s (backed by build: %s)\n", name, buildName)
	return c.buildTailer.Tail(ctx, writer, buildName, namespace)
}
