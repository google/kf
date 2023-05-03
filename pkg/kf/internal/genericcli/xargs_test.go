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
	fakedynamicclient "knative.dev/pkg/injection/clients/dynamicclient/fake"
)

func TestXargsCommand_docs(t *testing.T) {
	typ := &genericType{
		KubernetesType: KubernetesType{
			NsScoped: true, // only NsScoped types are supported
			Group:    "kf.dev",
			Version:  "v1alpha1",
			Kind:     "App",
			Resource: "apps",
			KfName:   "MyCustomType",
		},
	}

	cases := map[string]struct {
		genericType *genericType
		opts        []XargsOption

		wantUse     string
		wantShort   string
		wantLong    string
		wantAliases []string
		wantExample string
	}{
		"general": {
			genericType: typ,
			wantUse:     "xargs-mycustomtypes",
			wantShort:   "Run a command for every MyCustomType.",
			wantLong:    "Run a command for every MyCustomType in targeted spaces.",
			wantExample: "kf xargs-mycustomtypes",
		},
		"with custom example": {
			genericType: typ,
			wantUse:     "xargs-mycustomtypes",
			wantShort:   "Run a command for every MyCustomType.",
			wantLong:    "Run a command for every MyCustomType in targeted spaces.",
			wantExample: "a custom example",
			opts: []XargsOption{
				WithXargsExample("a custom example"),
			},
		},
		"with aliases": {
			genericType: typ,
			wantUse:     "xargs-mycustomtypes",
			wantShort:   "Run a command for every MyCustomType.",
			wantLong:    "Run a command for every MyCustomType in targeted spaces.",
			wantExample: "kf xargs-mycustomtypes",
			wantAliases: []string{"a", "b"},
			opts: []XargsOption{
				WithXargsAliases([]string{"a", "b"}),
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			cmd := NewXargsCommand(tc.genericType, nil, nil, tc.opts...)
			testutil.AssertEqual(t, "use", tc.wantUse, cmd.Use)
			testutil.AssertEqual(t, "short", tc.wantShort, cmd.Short)
			testutil.AssertEqual(t, "long", tc.wantLong, cmd.Long)
			testutil.AssertEqual(t, "aliases", tc.wantAliases, cmd.Aliases)
			testutil.AssertEqual(t, "example", tc.wantExample, cmd.Example)
		})
	}
}

func TestNewXargsCommand(t *testing.T) {
	nsType := &genericType{
		KubernetesType: KubernetesType{
			NsScoped: true,
			Group:    "apps",
			Version:  "v1",
			Kind:     "Deployment",
			Resource: "deployments",
			KfName:   "App",
		},
	}

	clusterType := &genericType{
		KubernetesType: KubernetesType{
			NsScoped: false,
			Group:    "apps",
			Version:  "v1",
			Kind:     "Deployment",
			Resource: "deployments",
			KfName:   "App",
		},
	}

	type mocks struct {
		p    *config.KfParams
		opts []XargsOption
	}

	type resource struct {
		Ns   string
		Name string
	}

	cases := map[string]struct {
		t       *genericType
		args    []string
		setup   func(context.Context, *testing.T, *mocks)
		wantOut string
		wantErr error
		create  []resource
	}{
		"wrong number of params": {
			t:       nsType,
			args:    []string{},
			wantErr: errors.New("requires at least 1 arg(s), only received 0"),
		},
		"namespace no target": {
			t:    nsType,
			args: []string{"noop"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.p.Space = ""
			},
			wantErr: errors.New(config.EmptySpaceError),
		},
		"namespace type": {
			t:    nsType,
			args: []string{"{{.Space}} {{.Name}}"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.p.Space = "my-ns"
				obj := nsType.NewUnstructured("my-ns", "some-object-name")
				fakedynamicclient.Get(ctx).
					Resource(nsType.GroupVersionResource(context.Background())).
					Namespace("my-ns").
					Create(ctx, obj, metav1.CreateOptions{})
			},
			wantOut: "# Xargs Apps in space: my-ns\n# Command for space=my-ns App=some-object-name\n'my-ns some-object-name'\n# Run with --dry-run=false to apply.\n",
		},
		"cluster type": {
			t:    clusterType,
			args: []string{"{{.Space}} {{.Name}}"},
			create: []resource{
				{
					Ns:   "my-ns",
					Name: "some-object-name",
				},
			},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.p.Space = "my-ns"
			},
			wantOut: "# Command for App=some-object-name\n' some-object-name'\n# Run with --dry-run=false to apply.\n",
		},
		"namespace type two namespaces four resources": {
			t:    nsType,
			args: []string{"{{.Space}} {{.Name}}"},
			create: []resource{
				{
					Ns:   "ns1",
					Name: "app1",
				},
				{
					Ns:   "ns1",
					Name: "app2",
				},
				{
					Ns:   "ns2",
					Name: "app3",
				},
				{
					Ns:   "ns2",
					Name: "app4",
				},
			},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.p.Space = "ns1,ns2"
			},
			wantOut: "# Xargs Apps in space: ns1,ns2\n# Command for space=ns1 App=app1\n'ns1 app1'\n# Command for space=ns1 App=app2\n'ns1 app2'\n# Command for space=ns2 App=app3\n'ns2 app3'\n# Command for space=ns2 App=app4\n'ns2 app4'\n# Run with --dry-run=false to apply.\n",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			buf := new(bytes.Buffer)
			ctx := fakeinjection.WithInjection(context.Background(), t)
			ctx = configlogging.SetupLogger(ctx, buf)

			mocks := &mocks{
				p: &config.KfParams{
					Space: "my-ns",
				},
			}

			for _, r := range tc.create {
				obj := tc.t.NewUnstructured(r.Ns, r.Name)
				fakedynamicclient.Get(ctx).
					Resource(nsType.GroupVersionResource(context.Background())).
					Namespace(r.Ns).
					Create(ctx, obj, metav1.CreateOptions{})
			}

			if tc.setup != nil {
				tc.setup(ctx, t, mocks)
			}

			cmd := NewXargsCommand(tc.t, mocks.p, nil, mocks.opts...)
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
