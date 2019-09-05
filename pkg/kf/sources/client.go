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

package sources

import (
	"context"
	"errors"
	"fmt"
	"io"

	cv1alpha1 "github.com/google/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1"
)

// ClientExtension holds additional functions that should be exposed by client.
type ClientExtension interface {
	Tail(ctx context.Context, namespace, name string, writer io.Writer) error
	Status(namespace, name string) (bool, error)
}

// BuildTailer is implemented by github.com/knative/build/pkg/logs
type BuildTailer interface {
	Tail(ctx context.Context, out io.Writer, buildName, namespace string) error
}

type sourcesClient struct {
	coreClient

	buildTailer BuildTailer
}

// NewClient creates a new build client.
func NewClient(kclient cv1alpha1.SourcesGetter, buildTailer BuildTailer) Client {
	return &sourcesClient{
		coreClient: coreClient{
			kclient:      kclient,
			upsertMutate: MutatorList{},
		},
		buildTailer: buildTailer,
	}
}

// Status gets the status of the source with the given name by calling SourceStatus.
// If the source doesn't exist an error is returned.
func (c *sourcesClient) Status(namespace, name string) (bool, error) {
	bld, err := c.coreClient.Get(namespace, name)
	if err != nil {
		return true, err
	}

	return SourceStatus(*bld)
}

// Tail streams the build logs to a local writer.
func (c *sourcesClient) Tail(ctx context.Context, namespace, name string, writer io.Writer) error {
	bld, err := c.coreClient.Get(namespace, name)
	if err != nil {
		return err
	}

	buildName := bld.Status.BuildName
	if buildName == "" {
		return errors.New("The build hasn't started yet")
	}

	fmt.Fprintf(writer, "Logs for %s (backed by build: %s)\n", name, buildName)
	return c.buildTailer.Tail(ctx, writer, buildName, namespace)
}
