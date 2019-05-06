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
	"fmt"

	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	logs "github.com/knative/build/pkg/logs"
	"github.com/segmentio/textio"

	v1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClientInterface is the main interface for interacting with builds.
type ClientInterface interface {
	Create(name string, template build.TemplateInstantiationSpec, opts ...CreateOption) (*build.Build, error)
	Logs(name string, opts ...LogsOption) error
	Status(name string, opts ...StatusOption) (complete bool, err error)
}

var _ ClientInterface = (*Client)(nil)

// Client is a client to knative.Build built in a way that other systems could
// be mostly dropped in as replacements.
type Client struct {
	buildClient cbuild.BuildV1alpha1Interface
}

// Create creates a new build.
func (c *Client) Create(name string, template build.TemplateInstantiationSpec, opts ...CreateOption) (*build.Build, error) {
	cfg := CreateOptionDefaults().Extend(opts).toConfig()

	// XXX: other types of sources are supported, notably git.
	// Adding in flags for git could be a quick win to support git-ops workflows.
	var source *build.SourceSpec
	if cfg.SourceImage != "" {
		source = &build.SourceSpec{
			Custom: &corev1.Container{
				Image: cfg.SourceImage,
			},
		}
	}

	args := []build.ArgumentSpec{}
	for k, v := range cfg.Args {
		args = append(args, build.ArgumentSpec{
			Name:  k,
			Value: v,
		})
	}

	build := &build.Build{
		Spec: build.BuildSpec{
			ServiceAccountName: cfg.ServiceAccount,
			Source:             source,
			Template: &build.TemplateInstantiationSpec{
				Name:      template.Name,
				Kind:      template.Kind,
				Arguments: args,
			},
		},
	}
	build.APIVersion = v1alpha1.SchemeGroupVersion.String()
	build.Kind = "Build"
	build.Name = name

	return c.buildClient.Builds(cfg.Namespace).Create(build)
}

// Logs tails the logs from the pod(s) running the build.
func (c *Client) Logs(name string, opts ...LogsOption) error {
	cfg := LogsOptionDefaults().Extend(opts).toConfig()

	p := textio.NewPrefixWriter(cfg.Output, fmt.Sprintf("[build %s] ", name))
	defer p.Flush()

	// XXX: This function uses its own loaders to grab credentials to connect to
	// Kubernetes rather than re-using the client cred. This is less than ideal,
	// but it's better to call out here than re-implement the logic and tightly
	// couple ourselves to Knative build.
	//
	// https://github.com/knative/build/blob/master/pkg/logs/logs.go
	return logs.Tail(context.Background(), p, name, cfg.Namespace)
}

// Status gets the status of the build with the given name.
// Complete will be set to true if the build has completed (or doesn't exist).
// Error will be set if the build completed with an error (or doesn't exist).
// A successful result is one that completed and error is nil.
func (c *Client) Status(name string, opts ...StatusOption) (bool, error) {
	bld, err := c.buildClient.Builds("default").Get(name, v1.GetOptions{})
	if err != nil {
		return true, fmt.Errorf("couldn't get build %q, %s", name, err.Error())
	}

	condition := bld.Status.GetCondition(duckv1alpha1.ConditionSucceeded)
	if condition == nil {
		// no success condition means the build is still going
		return false, nil
	}

	if condition.IsFalse() {
		// the build was a failure
		return true, fmt.Errorf("build failed for reason: %s with message: %s", condition.Reason, condition.Message)
	}

	// the build succeeded
	return true, nil
}
