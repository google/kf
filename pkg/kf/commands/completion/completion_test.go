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

package completion

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	fakedynamicclient "knative.dev/pkg/injection/clients/dynamicclient/fake"
)

func newKfObject(kind, name, ns string) runtime.Object {
	return newObject("kf.dev/v1alpha1", kind, name, ns)
}

func newObject(apiVersion, kind, name, ns string) runtime.Object {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      name,
			},
		},
	}
}

func TestGenericCompletionFn(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ToComplete    string
		Want          []string
		Resource      schema.GroupVersionResource
		GvrToListKind map[schema.GroupVersionResource]string
		Objects       []runtime.Object
		Namespace     string
	}{
		"namespaced resource": {
			Namespace: "my-ns",
			Resource: schema.GroupVersionResource{
				Group:    "kf.dev",
				Version:  "v1alpha1",
				Resource: "apps",
			},
			GvrToListKind: map[schema.GroupVersionResource]string{
				{Group: "kf.dev", Version: "v1alpha1", Resource: "apps"}: "AppList",
			},
			Objects: []runtime.Object{
				newKfObject("App", "foo", "my-ns"),
				newKfObject("App", "bar", "my-ns"),
				newKfObject("App", "baz", "my-ns"),
				newKfObject("App", "bull", "other-ns"),
			},
			ToComplete: "b",
			Want:       []string{"bar", "baz"},
		},
		"cluster resource": {
			Resource: schema.GroupVersionResource{
				Group:    "kf.dev",
				Version:  "v1alpha1",
				Resource: "spaces",
			},
			GvrToListKind: map[schema.GroupVersionResource]string{
				{Group: "kf.dev", Version: "v1alpha1", Resource: "spaces"}: "SpaceList",
			},
			Objects: []runtime.Object{
				newKfObject("Space", "foo", ""),
				newKfObject("Space", "bar", ""),
				newKfObject("Space", "for", ""),
			},
			ToComplete: "fo",
			Want:       []string{"foo", "for"},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {

			scheme := runtime.NewScheme()

			objects, err := convertObjectsToUnstructured(scheme, tc.Objects)
			if err != nil {
				panic(err)
			}

			for _, obj := range objects {
				gvk := obj.GetObjectKind().GroupVersionKind()
				if !scheme.Recognizes(gvk) {
					scheme.AddKnownTypeWithName(gvk, &unstructured.Unstructured{})
				}
				gvk.Kind += "List"
				if !scheme.Recognizes(gvk) {
					scheme.AddKnownTypeWithName(gvk, &unstructured.UnstructuredList{})
				}
			}

			client := fake.NewSimpleDynamicClientWithCustomListKinds(
				scheme,
				tc.GvrToListKind,
				tc.Objects...)
			var dynClient dynamic.ResourceInterface
			if tc.Namespace != "" {
				dynClient = client.Resource(tc.Resource).Namespace(tc.Namespace)
			} else {
				dynClient = client.Resource(tc.Resource)
			}
			completions, shellCompDirective := GenericCompletionFn(dynClient)(&cobra.Command{}, []string{}, tc.ToComplete)

			testutil.AssertEqual(t, "err", cobra.ShellCompDirectiveNoFileComp, shellCompDirective)
			testutil.AssertEqual(t, "completions", tc.Want, completions)
		})
	}
}

func TestAppCompletionFn(t *testing.T) {
	ctx, _ := fakedynamicclient.With(
		context.Background(),
		runtime.NewScheme(),
		newKfObject("App", "foo", "my-space"),
		newKfObject("App", "fi", "my-space"),
		newKfObject("App", "bar", "my-space"),
		newKfObject("App", "fum", "other-space"),
	)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	p := &config.KfParams{
		Space: "my-space",
	}
	completions, err := AppCompletionFn(p)(cmd, []string{}, "f")
	testutil.AssertEqual(t, "err", cobra.ShellCompDirectiveNoFileComp, err)
	testutil.AssertEqual(t, "completions", []string{"fi", "foo"}, completions)
}

func TestBuildCompletionFn(t *testing.T) {
	ctx, _ := fakedynamicclient.With(
		context.Background(),
		runtime.NewScheme(),
		newKfObject("Build", "foo", "my-space"),
		newKfObject("Build", "fi", "my-space"),
		newKfObject("Build", "bar", "my-space"),
		newKfObject("Build", "fum", "other-space"),
	)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	p := &config.KfParams{
		Space: "my-space",
	}
	completions, err := BuildCompletionFn(p)(cmd, []string{}, "f")
	testutil.AssertEqual(t, "err", cobra.ShellCompDirectiveNoFileComp, err)
	testutil.AssertEqual(t, "completions", []string{"fi", "foo"}, completions)
}

func TestSpaceCompletionFn(t *testing.T) {
	ctx, _ := fakedynamicclient.With(
		context.Background(),
		runtime.NewScheme(),
		newKfObject("Space", "foo", ""),
		newKfObject("Space", "fi", ""),
		newKfObject("Space", "bar", ""),
		newKfObject("Space", "fum", ""),
	)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	p := &config.KfParams{
		Space: "my-space",
	}
	completions, err := SpaceCompletionFn(p)(cmd, []string{}, "f")
	testutil.AssertEqual(t, "err", cobra.ShellCompDirectiveNoFileComp, err)
	testutil.AssertEqual(t, "completions", []string{"fi", "foo", "fum"}, completions)
}

func TestNetworkPolicyCompletionFn(t *testing.T) {
	ctx, _ := fakedynamicclient.With(
		context.Background(),
		runtime.NewScheme(),
		newObject("networking.k8s.io/v1", "NetworkPolicy", "foo", "my-space"),
		newObject("networking.k8s.io/v1", "NetworkPolicy", "fi", "my-space"),
		newObject("networking.k8s.io/v1", "NetworkPolicy", "bar", "my-space"),
		newObject("networking.k8s.io/v1", "NetworkPolicy", "fum", "other-space"),
	)
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	p := &config.KfParams{
		Space: "my-space",
	}
	completions, err := NetworkPolicyCompletionFn(p)(cmd, []string{}, "f")
	testutil.AssertEqual(t, "err", cobra.ShellCompDirectiveNoFileComp, err)
	testutil.AssertEqual(t, "completions", []string{"fi", "foo"}, completions)
}

func convertObjectsToUnstructured(s *runtime.Scheme, objs []runtime.Object) ([]runtime.Object, error) {
	ul := make([]runtime.Object, 0, len(objs))

	for _, obj := range objs {
		u, err := convertToUnstructured(s, obj)
		if err != nil {
			return nil, err
		}

		ul = append(ul, u)
	}
	return ul, nil
}

func convertToUnstructured(s *runtime.Scheme, obj runtime.Object) (runtime.Object, error) {
	var (
		err error
		u   unstructured.Unstructured
	)

	u.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to unstructured: %w", err)
	}

	gvk := u.GroupVersionKind()
	if gvk.Group == "" || gvk.Kind == "" {
		gvks, _, err := s.ObjectKinds(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to unstructured - unable to get GVK %w", err)
		}
		apiv, k := gvks[0].ToAPIVersionAndKind()
		u.SetAPIVersion(apiv)
		u.SetKind(k)
	}
	return &u, nil
}
