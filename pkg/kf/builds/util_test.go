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

package builds_test

import (
	"errors"
	"testing"

	"github.com/google/kf/pkg/kf/builds"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/knative/build/pkg/apis/build/v1alpha1"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildStatus(t *testing.T) {
	cases := map[string]struct {
		build          build.Build
		expectFinished bool
		expectErr      error
	}{
		"incomplete": {
			build:          build.Build{},
			expectFinished: false,
			expectErr:      nil,
		},
		"failed": {
			build: build.Build{
				Status: build.BuildStatus{
					Status: duckv1alpha1.Status{
						Conditions: duckv1alpha1.Conditions{
							{Type: duckv1alpha1.ConditionSucceeded, Status: "False", Reason: "fail-reason", Message: "fail-message"},
						},
					},
				},
			},
			expectFinished: true,
			expectErr:      errors.New("build failed for reason: fail-reason with message: fail-message"),
		},
		"succeeded": {
			build: build.Build{
				Status: build.BuildStatus{
					Status: duckv1alpha1.Status{
						Conditions: duckv1alpha1.Conditions{
							{Type: duckv1alpha1.ConditionSucceeded, Status: corev1.ConditionTrue},
						},
					},
				},
			},
			expectFinished: true,
			expectErr:      nil,
		},
		"still building": {
			build: build.Build{
				Status: build.BuildStatus{
					Status: duckv1alpha1.Status{
						Conditions: duckv1alpha1.Conditions{
							{Type: duckv1alpha1.ConditionSucceeded, Status: corev1.ConditionUnknown, Reason: "Building"},
						},
					},
				},
			},
			expectFinished: false,
			expectErr:      nil,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			finished, err := builds.BuildStatus(tc.build)

			testutil.AssertEqual(t, "finished", tc.expectFinished, finished)
			testutil.AssertErrorsEqual(t, tc.expectErr, err)
		})
	}
}

func TestPopulateTemplate(t *testing.T) {
	testTemplate := build.TemplateInstantiationSpec{
		Name: "buildpack",
		Kind: "ClusterBuildTemplate",
	}

	cases := map[string]struct {
		name     string
		template build.TemplateInstantiationSpec
		opts     []builds.CreateOption

		validate func(t *testing.T, actual *build.Build)
	}{
		"metadata": {
			name:     "vkn",
			template: testTemplate,
			opts: []builds.CreateOption{
				builds.WithCreateOwner(&v1.OwnerReference{
					UID: "abcd-efgh-ijkl",
				}),
			},

			validate: func(t *testing.T, actual *build.Build) {
				testutil.AssertEqual(t, "SchemeGroupVersion", build.SchemeGroupVersion.String(), actual.APIVersion)
				testutil.AssertEqual(t, "Kind", "Build", actual.Kind)
				testutil.AssertEqual(t, "Name", "vkn", actual.Name)
				testutil.AssertEqual(t, "owner count", 1, len(actual.OwnerReferences))
				testutil.AssertEqual(t, "owner values", "abcd-efgh-ijkl", string(actual.OwnerReferences[0].UID))
			},
		},
		"defaults": {
			name:     "defaults",
			template: testTemplate,
			opts:     []builds.CreateOption{},

			validate: func(t *testing.T, actual *build.Build) {
				testutil.AssertEqual(t, "Namespace", "default", actual.Namespace)

				spec := actual.Spec
				testutil.AssertEqual(t, "Spec.ServiceAccountName", "", spec.ServiceAccountName)

				source := spec.Source
				testutil.AssertEqual(t, "Spec.Source", (*v1alpha1.SourceSpec)(nil), source)

				template := spec.Template
				testutil.AssertEqual(t, "Spec.Template.Name", "buildpack", template.Name)
				testutil.AssertEqual(t, "Spec.Template.Kind", "ClusterBuildTemplate", string(template.Kind))
				testutil.AssertEqual(t, "Spec.Template.Arguments", []build.ArgumentSpec{}, template.Arguments)
			},
		},
		"custom": {
			name: "custom",
			template: build.TemplateInstantiationSpec{
				Name: "custom-template",
				Kind: "BuildTemplate",
			},
			opts: []builds.CreateOption{
				builds.WithCreateServiceAccount("sa"),
				builds.WithCreateSourceImage("oci.hub/source"),
				builds.WithCreateArgs(map[string]string{"FOO": "BAR"}),
				builds.WithCreateNamespace("custom-ns"),
			},

			validate: func(t *testing.T, actual *build.Build) {
				testutil.AssertEqual(t, "Namespace", "custom-ns", actual.Namespace)

				spec := actual.Spec
				testutil.AssertEqual(t, "Spec.ServiceAccountName", "sa", spec.ServiceAccountName)

				source := spec.Source
				testutil.AssertEqual(t, "Spec.Source", "oci.hub/source", source.Custom.Image)

				template := spec.Template
				testutil.AssertEqual(t, "Spec.Template.Name", "custom-template", template.Name)
				testutil.AssertEqual(t, "Spec.Template.Kind", "BuildTemplate", string(template.Kind))
				testutil.AssertEqual(t, "Spec.Template.Arguments", []build.ArgumentSpec{
					{Name: "FOO", Value: "BAR"},
				}, template.Arguments)
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			out := builds.PopulateTemplate(tc.name, tc.template, tc.opts...)
			tc.validate(t, out)
		})
	}
}
