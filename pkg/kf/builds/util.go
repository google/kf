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
	"io"

	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildStatus gets the status of the given build.
// Complete will be set to true if the build has completed (or doesn't exist).
// Error will be set if the build completed with an error (or doesn't exist).
// A successful result is one that completed and error is nil.
func BuildStatus(build build.Build) (finished bool, err error) {
	condition := build.Status.GetCondition(duckv1alpha1.ConditionSucceeded)
	if condition == nil {
		// no success condition means the build hasn't propigated yet
		return false, nil
	}

	switch condition.Status {
	case corev1.ConditionTrue: // the build was a success
		return true, nil

	case corev1.ConditionFalse: // the build was a failure
		return true, fmt.Errorf("build failed for reason: %s with message: %s", condition.Reason, condition.Message)

	default: // the build is in a transition state
		return false, nil
	}
}

// PopulateTemplate populates a build template for Client.Create.
// This method is public for debugging purposes so we can dry-run builds for
// users.
func PopulateTemplate(
	name string,
	template build.TemplateInstantiationSpec,
	opts ...CreateOption,
) *build.Build {
	cfg := CreateOptionDefaults().Extend(opts).toConfig()

	// XXX: other types of sources are supported, notably git.
	// Adding in flags for git could be a quick win to support git-ops
	// workflows.
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

	out := &build.Build{
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
	out.APIVersion = build.SchemeGroupVersion.String()
	out.Kind = "Build"
	out.Name = name
	out.Namespace = cfg.Namespace

	if cfg.Owner != nil {
		out.OwnerReferences = []v1.OwnerReference{*cfg.Owner}
	}

	return out
}

// BuildTailerFunc converts a func into a BuildTailer.
type BuildTailerFunc func(ctx context.Context, out io.Writer, buildName, namespace string) error

// Tail implements BuildTailer.
func (f BuildTailerFunc) Tail(ctx context.Context, out io.Writer, buildName, namespace string) error {
	return f(ctx, out, buildName, namespace)
}
