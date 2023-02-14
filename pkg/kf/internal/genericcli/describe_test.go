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

package genericcli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	fakeinjection "github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	fakedynamicclient "knative.dev/pkg/injection/clients/dynamicclient/fake"
)

type genericType struct {
	KubernetesType
}

func (g *genericType) NewUnstructured(ns, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": g.Group + "/" + g.Version,
			"kind":       g.Kind,
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      name,
			},
		},
	}
}

func (g *genericType) NewTable(name string) *metav1.Table {
	return &metav1.Table{
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{Name: "Name"},
		},
		Rows: []metav1.TableRow{
			{Cells: []interface{}{name}},
		},
	}
}

func TestNewDescribeCommand_docs(t *testing.T) {
	typ := &genericType{
		KubernetesType: KubernetesType{
			NsScoped: true,
			Group:    "kf.dev",
			Version:  "v1alpha1",
			Kind:     "App",
			Resource: "apps",
			KfName:   "KfApp",
		},
	}

	cases := map[string]struct {
		genericType *genericType
		opts        []DescribeOption

		wantUse     string
		wantShort   string
		wantLong    string
		wantAliases []string
	}{
		"general": {
			genericType: typ,

			wantUse:   "kfapp NAME",
			wantShort: "Print information about the given KfApp.",
			wantLong:  "Print information about the given KfApp.",
		},

		"overrides": {
			genericType: typ,
			opts: []DescribeOption{
				WithDescribeCommandName("some-name"),
				WithDescribeAliases([]string{"abc"}),
			},
			wantUse:     "some-name NAME",
			wantShort:   "Print information about the given KfApp.",
			wantLong:    "Print information about the given KfApp.",
			wantAliases: []string{"abc"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			cmd := NewDescribeCommand(tc.genericType, nil, tc.opts...)
			testutil.AssertEqual(t, "use", tc.wantUse, cmd.Use)
			testutil.AssertEqual(t, "short", tc.wantShort, cmd.Short)
			testutil.AssertEqual(t, "long", tc.wantLong, cmd.Long)
			testutil.AssertEqual(t, "aliases", tc.wantAliases, cmd.Aliases)
		})
	}
}

func TestNewDescribeCommand(t *testing.T) {
	nsType := &genericType{
		KubernetesType: KubernetesType{
			NsScoped: true,
			Group:    "kf.dev",
			Version:  "v1alpha1",
			Kind:     "App",
			Resource: "apps",
			KfName:   "App",
		},
	}

	clusterType := &genericType{
		KubernetesType: KubernetesType{
			NsScoped: false,
			Group:    "kf.dev",
			Version:  "v1alpha1",
			Kind:     "Space",
			Resource: "spaces",
			KfName:   "Space",
		},
	}

	type mocks struct {
		p *config.KfParams
	}

	cases := map[string]struct {
		t       Type
		args    []string
		setup   func(context.Context, *testing.T, *mocks)
		wantOut string
		wantErr error
	}{
		"no params": {
			t:       nsType,
			args:    []string{},
			wantErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"namespace no target": {
			t:    nsType,
			args: []string{"name"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.p.Space = ""
			},
			wantErr: errors.New(config.EmptySpaceError),
		},
		"cluster good": {
			t:    clusterType,
			args: []string{"some-object-name"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				obj := clusterType.NewUnstructured("", "some-object-name")
				fakedynamicclient.Get(ctx).
					Resource(clusterType.GroupVersionResource(context.Background())).
					Create(ctx, obj, metav1.CreateOptions{})
			},
			wantOut: `Getting Space some-object-name
API Version:  kf.dev/v1alpha1
Kind:         Space
Metadata:
  Name:  some-object-name
`,
		},
		"missing object": {
			t:       clusterType,
			args:    []string{"some-object-name"},
			wantErr: errors.New(`spaces.kf.dev "some-object-name" not found`),
		},

		"namespace good": {
			t:    nsType,
			args: []string{"some-object-name"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.p.Space = "my-ns"
				obj := nsType.NewUnstructured("my-ns", "some-object-name")
				fakedynamicclient.Get(ctx).
					Resource(nsType.GroupVersionResource(context.Background())).
					Namespace("my-ns").
					Create(ctx, obj, metav1.CreateOptions{})
			},
			wantOut: `Getting App some-object-name in Space: my-ns
API Version:  kf.dev/v1alpha1
Kind:         App
Metadata:
  Name:       some-object-name
  Namespace:  my-ns
`,
		},
		"custom output": {
			t:    clusterType,
			args: []string{"some-object-name", "-o", "name"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				obj := clusterType.NewUnstructured("", "some-object-name")
				fakedynamicclient.Get(ctx).
					Resource(clusterType.GroupVersionResource(context.Background())).
					Create(ctx, obj, metav1.CreateOptions{})
			},
			wantOut: `Getting Space some-object-name
space.kf.dev/some-object-name
`,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			buf := new(bytes.Buffer)
			ctx := fakeinjection.WithInjection(context.Background(), t)
			ctx = configlogging.SetupLogger(ctx, buf)

			mocks := &mocks{
				p: &config.KfParams{
					Space: "default",
				},
			}

			if tc.setup != nil {
				tc.setup(ctx, t, mocks)
			}

			cmd := NewDescribeCommand(tc.t, mocks.p)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)

			_, actualErr := cmd.ExecuteC()
			if tc.wantErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
				return
			}

			testutil.AssertEqual(t, "output", buf.String(), tc.wantOut)
		})
	}
}
