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

	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/testutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakedynamic "k8s.io/client-go/dynamic/fake"
)

type genericType struct {
	NsScoped bool
	Group    string
	Version  string
	Kind     string
	Resource string
	KfName   string
}

var _ Type = (*genericType)(nil)

func (g *genericType) Namespaced() bool {
	return g.NsScoped
}

func (g *genericType) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    g.Group,
		Version:  g.Version,
		Resource: g.Resource,
	}
}

func (g *genericType) GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   g.Group,
		Version: g.Version,
		Kind:    g.Kind,
	}
}

func (g *genericType) FriendlyName() string {
	return g.KfName
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

func (g *genericType) NewTable(name string) *metav1beta1.Table {
	return &metav1beta1.Table{
		ColumnDefinitions: []metav1beta1.TableColumnDefinition{
			{Name: "Name"},
		},
		Rows: []metav1beta1.TableRow{
			{Cells: []interface{}{name}},
		},
	}
}

func TestNewDescribeCommand_docs(t *testing.T) {
	typ := &genericType{
		NsScoped: true,
		Group:    "kf.dev",
		Version:  "v1alpha1",
		Kind:     "App",
		Resource: "apps",
		KfName:   "KfApp",
	}

	cmd := NewDescribeCommand(typ, nil, nil)
	testutil.AssertEqual(t, "use", "kfapp NAME", cmd.Use)
	testutil.AssertEqual(t, "short", "Print information about the given KfApp", cmd.Short)
	testutil.AssertEqual(t, "long", "Print information about the given KfApp", cmd.Long)
}

func TestNewDescribeCommand(t *testing.T) {
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
		p      *config.KfParams
		client *fakedynamic.FakeDynamicClient
	}

	cases := map[string]struct {
		t       Type
		args    []string
		setup   func(*testing.T, *mocks)
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
			setup: func(t *testing.T, mocks *mocks) {
				mocks.p.Namespace = ""
			},
			wantErr: errors.New("no space targeted, use 'kf target --space SPACE' to target a space"),
		},
		"cluster good": {
			t:    clusterType,
			args: []string{"some-object-name"},
			setup: func(t *testing.T, mocks *mocks) {
				obj := clusterType.NewUnstructured("", "some-object-name")
				mocks.client = fakedynamic.NewSimpleDynamicClient(runtime.NewScheme(), obj)
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
			setup: func(t *testing.T, mocks *mocks) {
				mocks.p.Namespace = "my-ns"
				obj := nsType.NewUnstructured("my-ns", "some-object-name")
				mocks.client = fakedynamic.NewSimpleDynamicClient(runtime.NewScheme(), obj)
			},
			wantOut: `Getting App some-object-name in namespace: my-ns
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
			setup: func(t *testing.T, mocks *mocks) {
				obj := clusterType.NewUnstructured("", "some-object-name")
				mocks.client = fakedynamic.NewSimpleDynamicClient(runtime.NewScheme(), obj)
			},
			wantOut: `Getting Space some-object-name
space.kf.dev/some-object-name
`,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			mocks := &mocks{
				p: &config.KfParams{
					Namespace: "default",
				},
				client: fakedynamic.NewSimpleDynamicClient(runtime.NewScheme()),
			}

			if tc.setup != nil {
				tc.setup(t, mocks)
			}

			buf := new(bytes.Buffer)
			cmd := NewDescribeCommand(tc.t, mocks.p, mocks.client)
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
