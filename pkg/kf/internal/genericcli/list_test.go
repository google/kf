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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/commands/config"
	faketable "github.com/google/kf/pkg/kf/internal/tableclient/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"k8s.io/apimachinery/pkg/runtime"
	fakedynamic "k8s.io/client-go/dynamic/fake"
)

func TestNewListCommand_docs(t *testing.T) {
	typ := &genericType{
		NsScoped: true,
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Kind:     "App",
		Resource: "apps",
		KfName:   "KfApp",
	}

	cmd := NewListCommand(typ, nil, nil, nil)
	testutil.AssertEqual(t, "use", "kfapps", cmd.Use)
	testutil.AssertEqual(t, "short", "List KfApps in the target space", cmd.Short)
	testutil.AssertEqual(t, "long", "List KfApps in the target space", cmd.Long)
}

func TestNewListCommand(t *testing.T) {
	nsType := &genericType{
		NsScoped: true,
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Kind:     "App",
		Resource: "apps",
		KfName:   "App",
	}

	clusterType := &genericType{
		NsScoped: false,
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Kind:     "Space",
		Resource: "spaces",
		KfName:   "Space",
	}

	type mocks struct {
		p           *config.KfParams
		client      *fakedynamic.FakeDynamicClient
		tableClient *faketable.FakeInterface
	}

	cases := map[string]struct {
		t       Type
		args    []string
		setup   func(*testing.T, *mocks)
		wantOut string
		wantErr error
	}{
		"wrong number of params": {
			t:       nsType,
			args:    []string{"some", "params", "here"},
			wantErr: errors.New("accepts 0 arg(s), received 3"),
		},
		"namespace no target": {
			t:    nsType,
			args: []string{},
			setup: func(t *testing.T, mocks *mocks) {
				mocks.p.Namespace = ""
			},
			wantErr: errors.New("no space targeted, use 'kf target --space SPACE' to target a space"),
		},
		"cluster good": {
			t:    clusterType,
			args: []string{},
			setup: func(t *testing.T, mocks *mocks) {
				obj := clusterType.NewTable("some-object-name")
				mocks.tableClient.EXPECT().Table(gomock.Any(), gomock.Any(), gomock.Any()).Return(obj, nil)
			},
			wantOut: `Listing Spaces
Name
some-object-name
`,
		},
		"namespace good": {
			t:    nsType,
			args: []string{},
			setup: func(t *testing.T, mocks *mocks) {
				mocks.p.Namespace = "my-ns"
				obj := clusterType.NewTable("some-object-name")
				mocks.tableClient.EXPECT().Table(gomock.Any(), "my-ns", gomock.Any()).Return(obj, nil)
			},
			wantOut: `Listing Apps in namespace: my-ns
Name
some-object-name
`,
		},
		"custom output cluster": {
			t:    clusterType,
			args: []string{"-o", "name"},
			setup: func(t *testing.T, mocks *mocks) {
				obj := clusterType.NewUnstructured("", "some-object-name")
				mocks.client = fakedynamic.NewSimpleDynamicClient(runtime.NewScheme(), obj)
			},
			wantOut: `Listing Spaces
space.kf.dev/some-object-name
`,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mocks := &mocks{
				p: &config.KfParams{
					Namespace: "default",
				},
				client:      fakedynamic.NewSimpleDynamicClient(runtime.NewScheme()),
				tableClient: faketable.NewFakeInterface(ctrl),
			}

			if tc.setup != nil {
				tc.setup(t, mocks)
			}

			buf := new(bytes.Buffer)
			cmd := NewListCommand(tc.t, mocks.p, mocks.client, mocks.tableClient)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)

			_, actualErr := cmd.ExecuteC()
			if tc.wantErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
				return
			}

			testutil.AssertEqual(t, "output", buf.String(), tc.wantOut)
		})
	}
}
