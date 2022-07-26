// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/MakeNowJust/heredoc"
	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	fakeinjection "github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakedynamicclient "knative.dev/pkg/injection/clients/dynamicclient/fake"
)

func TestNewDeleteByNameCommand_docs(t *testing.T) {
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
		opts        []DeleteByNameOption

		wantUse     string
		wantShort   string
		wantLong    string
		wantAliases []string
	}{
		"general": {
			genericType: typ,
			wantUse:     "delete-kfapp NAME",
			wantShort:   "Delete the KfApp with the given name in the targeted Space.",
			wantLong: heredoc.Doc(`Deletes the KfApp with the given name and wait for it to be deleted.

    Kubernetes will delete the KfApp once all child resources it owns have been deleted.
    Deletion may take a long time if any of the following conditions are true:

    * There are many child objects.
    * There are finalizers on the object preventing deletion.
    * The cluster is in an unhealthy state.`),
		},

		"overrides": {
			genericType: typ,
			opts: []DeleteByNameOption{
				WithDeleteByNameCommandName("some-name"),
				WithDeleteByNameAliases([]string{"abc"}),
				WithDeleteByNameAdditionalLongText(`
					Some extra note here.

					* A
					* B`),
			},
			wantUse:   "some-name NAME",
			wantShort: "Delete the KfApp with the given name in the targeted Space.",
			wantLong: heredoc.Doc(`Deletes the KfApp with the given name and wait for it to be deleted.

    Kubernetes will delete the KfApp once all child resources it owns have been deleted.
    Deletion may take a long time if any of the following conditions are true:

    * There are many child objects.
    * There are finalizers on the object preventing deletion.
    * The cluster is in an unhealthy state.

    Some extra note here.

    * A
    * B`),
			wantAliases: []string{"abc"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			cmd := NewDeleteByNameCommand(tc.genericType, nil, tc.opts...)
			testutil.AssertEqual(t, "use", tc.wantUse, cmd.Use)
			testutil.AssertEqual(t, "short", tc.wantShort, cmd.Short)
			testutil.AssertEqual(t, "long", tc.wantLong, cmd.Long)
			testutil.AssertEqual(t, "aliases", tc.wantAliases, cmd.Aliases)
		})
	}
}

func TestNewDeleteByNameCommand(t *testing.T) {
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
		"wrong number of params": {
			t:       nsType,
			args:    []string{"some", "params", "here"},
			wantErr: errors.New("accepts 1 arg(s), received 3"),
		},
		"namespace no target": {
			t:    nsType,
			args: []string{"deleteme"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.p.Space = ""
			},
			wantErr: errors.New(config.EmptySpaceError),
		},
		"does-not-exist error": {
			t:    clusterType,
			args: []string{"does-not-exist"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				obj := clusterType.NewUnstructured("", "some-object-name")
				fakedynamicclient.Get(ctx).
					Resource(clusterType.GroupVersionResource(context.Background())).
					Create(ctx, obj, metav1.CreateOptions{})
			},
			wantErr: errors.New(`spaces.kf.dev "does-not-exist" not found`),
		},
		"cluster good": {
			t:    clusterType,
			args: []string{"deleteme"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				obj := clusterType.NewUnstructured("", "deleteme")
				fakedynamicclient.Get(ctx).
					Resource(clusterType.GroupVersionResource(context.Background())).
					Create(ctx, obj, metav1.CreateOptions{})
			},
			wantOut: `Deleting Space "deleteme"
Waiting for deletion...
`,
		},
		"namespaced good": {
			t:    nsType,
			args: []string{"deleteme"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.p.Space = "some-ns"
				obj := nsType.NewUnstructured("some-ns", "deleteme")
				fakedynamicclient.Get(ctx).
					Resource(nsType.GroupVersionResource(context.Background())).
					Namespace("some-ns").
					Create(ctx, obj, metav1.CreateOptions{})
			},
			wantOut: `Deleting App "deleteme" in Space "some-ns"
Waiting for deletion...
`,
		},
		"namespaced async": {
			t:    nsType,
			args: []string{"deleteme", "--async"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.p.Space = "some-ns"
				obj := nsType.NewUnstructured("some-ns", "deleteme")
				fakedynamicclient.Get(ctx).
					Resource(nsType.GroupVersionResource(context.Background())).
					Namespace("some-ns").
					Create(ctx, obj, metav1.CreateOptions{})
			},
			wantOut: `Deleting App "deleteme" in Space "some-ns"
`,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			gomock.NewController(t)

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

			cmd := NewDeleteByNameCommand(tc.t, mocks.p)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)

			_, actualErr := cmd.ExecuteC()
			if tc.wantErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
				return
			}

			testutil.AssertEqual(t, "output", tc.wantOut, buf.String())
		})
	}
}
