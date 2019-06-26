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
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/kf/pkg/kf/builds"
	"github.com/google/kf/pkg/kf/testutil"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	testclient "github.com/knative/build/pkg/client/clientset/versioned/fake"
	"github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClientCreate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		name     string
		template build.TemplateInstantiationSpec
		opts     []builds.CreateOption
		init     func(builds.ClientInterface, v1alpha1.BuildV1alpha1Interface)

		expectedErr       error
		expectedNamespace string
	}{
		"custom-namespace": {
			name:     "default",
			template: builds.BuildpackTemplate(),
			opts: []builds.CreateOption{
				builds.WithCreateNamespace("custom-ns"),
			},
			expectedNamespace: "custom-ns",
		},
		"default": {
			name:              "default",
			template:          builds.BuildpackTemplate(),
			opts:              []builds.CreateOption{},
			expectedNamespace: "default",
		},
		"already exists": {
			name:     "default",
			template: builds.BuildpackTemplate(),
			opts:     []builds.CreateOption{},
			init: func(client builds.ClientInterface, builder v1alpha1.BuildV1alpha1Interface) {
				client.Create("default", builds.BuildpackTemplate())
			},
			expectedErr: errors.New(`builds.build.knative.dev "default" already exists`),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			fakeClient := testclient.NewSimpleClientset().BuildV1alpha1()
			b := builds.NewClient(fakeClient, nil)

			if tc.init != nil {
				tc.init(b, fakeClient)
			}

			_, err := b.Create(tc.name, tc.template, tc.opts...)
			if err != nil || tc.expectedErr != nil {
				testutil.AssertErrorsEqual(t, tc.expectedErr, err)
				return
			}

			if _, err := fakeClient.Builds(tc.expectedNamespace).Get(tc.name, v1.GetOptions{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestClientStatus(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		name string
		opts []builds.StatusOption
		init func(builds.ClientInterface, v1alpha1.BuildV1alpha1Interface)

		expectedErr    error
		expectFinished bool
	}{
		"custom-namespace": {
			name: "default",
			init: func(client builds.ClientInterface, builder v1alpha1.BuildV1alpha1Interface) {
				client.Create("default", builds.BuildpackTemplate(), builds.WithCreateNamespace("custom-ns"))
			},
			opts: []builds.StatusOption{
				builds.WithStatusNamespace("custom-ns"),
			},
		},
		"default namespace": {
			name: "default",
			init: func(client builds.ClientInterface, builder v1alpha1.BuildV1alpha1Interface) {
				client.Create("default", builds.BuildpackTemplate(), builds.WithCreateNamespace("default"))
			},
			opts: []builds.StatusOption{},
		},
		"missing build": {
			name:           "missing",
			expectedErr:    errors.New(`couldn't get build "missing", builds.build.knative.dev "missing" not found`),
			expectFinished: true,
		},
		"finished build": {
			name: "completed",
			init: func(client builds.ClientInterface, builder v1alpha1.BuildV1alpha1Interface) {

				bld := &build.Build{}
				bld.Name = "completed"
				bld.Status.Conditions = duckv1alpha1.Conditions{
					{Type: duckv1alpha1.ConditionSucceeded, Status: corev1.ConditionTrue},
				}

				builder.Builds("default").Create(bld)

			},
			expectFinished: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			fakeClient := testclient.NewSimpleClientset().BuildV1alpha1()

			b := builds.NewClient(fakeClient, nil)

			if tc.init != nil {
				tc.init(b, fakeClient)
			}

			finished, err := b.Status(tc.name, tc.opts...)
			testutil.AssertErrorsEqual(t, tc.expectedErr, err)
			testutil.AssertEqual(t, "finished", tc.expectFinished, finished)
		})
	}
}

func TestClientDelete(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		name string
		opts []builds.DeleteOption
		init func(builds.ClientInterface, v1alpha1.BuildV1alpha1Interface)

		expectedErr    error
		expectFinished bool
	}{
		"custom-namespace": {
			name: "build1",
			init: func(client builds.ClientInterface, builder v1alpha1.BuildV1alpha1Interface) {
				client.Create("build1", builds.BuildpackTemplate(), builds.WithCreateNamespace("custom-ns"))
			},
			opts: []builds.DeleteOption{
				builds.WithDeleteNamespace("custom-ns"),
			},
		},
		"does-not-exist": {
			name:        "build1",
			expectedErr: errors.New(`builds.build.knative.dev "build1" not found`),
		},
		"defaults": {
			name: "build1",
			init: func(client builds.ClientInterface, builder v1alpha1.BuildV1alpha1Interface) {
				client.Create("build1", builds.BuildpackTemplate(), builds.WithCreateNamespace("default"))
			},
			opts: []builds.DeleteOption{},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			fakeClient := testclient.NewSimpleClientset().BuildV1alpha1()
			b := builds.NewClient(fakeClient, nil)

			if tc.init != nil {
				tc.init(b, fakeClient)
			}

			err := b.Delete(tc.name, tc.opts...)
			testutil.AssertErrorsEqual(t, tc.expectedErr, err)
		})
	}
}

func TestClientTail(t *testing.T) {
	// Tail proxies another function, so just make sure the params are passed in
	// the correct order.

	buildContext := context.TODO()
	buildWriter := &bytes.Buffer{}
	buildName := "some-build"
	buildNamespace := "some-namespace"

	tailerCalled := false
	asserter := builds.BuildTailerFunc(func(ctx context.Context, out io.Writer, name, ns string) error {
		tailerCalled = true

		testutil.AssertEqual(t, "context", buildContext, ctx)
		testutil.AssertEqual(t, "writer", buildWriter, out)
		testutil.AssertEqual(t, "build name", buildName, name)
		testutil.AssertEqual(t, "namespace", buildNamespace, ns)

		return errors.New("test-error")
	})

	// Make call
	b := builds.NewClient(nil, asserter)
	actualErr := b.Tail(
		buildName,
		builds.WithTailContext(buildContext),
		builds.WithTailWriter(buildWriter),
		builds.WithTailNamespace(buildNamespace))

	// Validate results
	testutil.AssertErrorsEqual(t, errors.New("test-error"), actualErr)
	testutil.AssertEqual(t, "tailer called", true, tailerCalled)
}
