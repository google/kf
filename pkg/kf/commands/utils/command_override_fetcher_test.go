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

package utils_test

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"testing"

	kf "github.com/GoogleCloudPlatform/kf/pkg/apis/kf/v1alpha1"
	"github.com/GoogleCloudPlatform/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1/fake"
	pkf "github.com/GoogleCloudPlatform/kf/pkg/kf"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/apps"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/utils"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	bfake "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1/fake"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestCommandOverrideFetcher(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Namespace          string
		BuildTailer        func(*testing.T) pkf.BuildTailer
		BuildImage         func(*testing.T) apps.SrcImageBuilder
		ListReactor        ktesting.ReactionFunc
		CreateBuildReactor ktesting.ReactionFunc
		DeleteBuildReactor ktesting.ReactionFunc
		Assert             func(t *testing.T, results map[string]*cobra.Command, err error)
	}{
		"properly fetches CommandSets": {
			Namespace: "some-namespace",
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				testutil.AssertEqual(t, "verb", "list", action.GetVerb())
				testutil.AssertEqual(t, "resource", "commandsets", action.GetResource().Resource)
				testutil.AssertEqual(t, "namespace", "some-namespace", action.GetNamespace())
				return false, nil, nil
			}),
		},
		"fetching CommandSets fails": {
			Namespace: "some-namespace",
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("some-error")
			}),
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertErrorsEqual(t, errors.New("fetching CommandSets failed: some-error"), err)
			},
		},
		"returns an error if multiple CommandSets are found": {
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{
					{}, {},
				}}, nil
			}),
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertErrorsEqual(t, errors.New("cluster is not properly setup. Expected a single CommandSet but have 2"), err)
			},
		},
		"uses name, short and long": {
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{
					{
						Spec: []kf.CommandSpec{
							{
								Name:  "some-name",
								Use:   "some-usage",
								Short: "some-short",
								Long:  "some-long",
							},
						},
					},
				}}, nil
			}),
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				c := results["some-name"]
				testutil.AssertNil(t, "error", err)
				testutil.AssertEqual(t, "use", "some-usage", c.Use)
				testutil.AssertEqual(t, "short", "some-short", c.Short)
				testutil.AssertEqual(t, "long", "some-long", c.Long)
			},
		},
		"Build error": {
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{{Spec: []kf.CommandSpec{{Name: "some-name"}}}}}, nil
			}),
			CreateBuildReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("some-error")
			}),
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertNil(t, "error", err)
				gotErr := results["some-name"].Execute()
				testutil.AssertErrorsEqual(t, errors.New("failed to create build: some-error"), gotErr)
			},
		},
		"build has namespace and generated name": {
			Namespace: "some-namespace",
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{{Spec: []kf.CommandSpec{{Name: "some-name"}}}}}, nil
			}),
			CreateBuildReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				testutil.AssertEqual(t, "namespace", "some-namespace", action.GetNamespace())
				testutil.AssertEqual(t, "verb", "create", action.GetVerb())
				testutil.AssertEqual(t, "resource", "builds", action.GetResource().Resource)
				b := action.(ktesting.CreateActionImpl).Object.(*build.Build)
				testutil.AssertEqual(t, "generated name", "kf-some-name", b.GenerateName)
				testutil.AssertEqual(t, "build namespace", "some-namespace", b.Namespace)

				return false, nil, nil
			}),
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertNil(t, "error", err)
				gotErr := results["some-name"].Execute()
				testutil.AssertNil(t, "error", gotErr)
			},
		},
		"Build spec has correct build template": {
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{{Spec: []kf.CommandSpec{{Name: "some-name", BuildTemplate: "some-template"}}}}}, nil
			}),
			CreateBuildReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				b := action.(ktesting.CreateActionImpl).Object.(*build.Build)
				testutil.AssertEqual(t, "build template", "some-template", b.Spec.Template.Name)

				return false, nil, nil
			}),
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertNil(t, "error", err)
				gotErr := results["some-name"].Execute()
				testutil.AssertNil(t, "error", gotErr)
			},
		},
		"Cleanup build": {
			Namespace: "some-namespace",
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{{Spec: []kf.CommandSpec{{Name: "some-name", BuildTemplate: "some-template"}}}}}, nil
			}),
			CreateBuildReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				b := action.(ktesting.CreateActionImpl).Object.(*build.Build)
				b.Name = "some-build"

				return true, b, nil
			}),
			DeleteBuildReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				testutil.AssertEqual(t, "namespace", "some-namespace", action.GetNamespace())
				testutil.AssertEqual(t, "verb", "delete", action.GetVerb())
				testutil.AssertEqual(t, "resource", "builds", action.GetResource().Resource)
				b := action.(ktesting.DeleteActionImpl)
				testutil.AssertEqual(t, "build name", "some-build", b.Name)

				return false, nil, nil
			}),
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertNil(t, "error", err)
				gotErr := results["some-name"].Execute()
				testutil.AssertNil(t, "error", gotErr)
			},
		},
		"debug-keep-build prevents build cleanup": {
			Namespace: "some-namespace",
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{{Spec: []kf.CommandSpec{{Name: "some-name", BuildTemplate: "some-template"}}}}}, nil
			}),
			CreateBuildReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				b := action.(ktesting.CreateActionImpl).Object.(*build.Build)
				b.Name = "some-build"

				return true, b, nil
			}),
			DeleteBuildReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				t.Fatal("this should not have happened")
				return false, nil, nil
			}),
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertNil(t, "error", err)
				c := results["some-name"]
				c.Flags().Set("debug-keep-build", "true")
				gotErr := c.Execute()
				testutil.AssertNil(t, "error", gotErr)
			},
		},
		"Build spec has correct flags": {
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{{Spec: []kf.CommandSpec{{
					Name:          "some-name",
					BuildTemplate: "some-template",
					Flags: []kf.Flag{
						{Type: "string", Long: "string-long", Short: "s", Default: "some-default", Description: "some-desc"},
						{Type: "stringArray", Long: "string-array-long", Short: "a", Description: "some-desc"},
						{Type: "bool", Long: "bool-long", Short: "b", Description: "some-desc"},
					},
				}}}}}, nil
			}),
			CreateBuildReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				b := action.(ktesting.CreateActionImpl).Object.(*build.Build)
				testutil.AssertEqual(t, "build flag name", "FLAGS", b.Spec.Template.Arguments[1].Name)
				testutil.AssertJSONEqual(t, `{"string-array-long":["[some-value-1,some-value-2]"],"string-long":["some-value"],"bool-long":["true"],"debug-keep-build":["true"]}`, b.Spec.Template.Arguments[1].Value)

				return false, nil, nil
			}),
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertNil(t, "error", err)
				c := results["some-name"]
				c.Flags().Set("string-long", "some-value")
				c.Flags().Set("string-array-long", "some-value-1")
				c.Flags().Set("string-array-long", "some-value-2")
				c.Flags().Set("bool-long", "true")
				c.Flags().Set("debug-keep-build", "true")
				gotErr := c.Execute()
				testutil.AssertNil(t, "error", gotErr)
			},
		},
		"Tail logs": {
			Namespace: "some-namespace",
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{{Spec: []kf.CommandSpec{{Name: "some-name", BuildTemplate: "some-template"}}}}}, nil
			}),
			CreateBuildReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				b := action.(ktesting.CreateActionImpl).Object.(*build.Build)
				b.Name = "some-build"

				return true, b, nil
			}),
			BuildTailer: func(t *testing.T) pkf.BuildTailer {
				return pkf.BuildTailerFunc(func(ctx context.Context, w io.Writer, name, namespace string) error {
					if ctx == nil {
						t.Fatal("expected ctx to not be nil")
					}
					testutil.AssertEqual(t, "name", "some-build", name)
					testutil.AssertEqual(t, "namespace", "some-namespace", namespace)
					return errors.New("some-error")
				})
			},
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertNil(t, "error", err)
				gotErr := results["some-name"].Execute()
				testutil.AssertErrorsEqual(t, errors.New("some-error"), gotErr)
			},
		},
		"upload dir without container registry": {
			Namespace: "some-namespace",
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{{
					ContainerRegistry: "",
					Spec:              []kf.CommandSpec{{Name: "some-name", BuildTemplate: "some-template", UploadDir: true}}}},
				}, nil
			}),
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertNil(t, "error", err)
				gotErr := results["some-name"].Execute()
				testutil.AssertErrorsEqual(t, errors.New("CRD failed to configure ContainerRegistry. Please contact your operator."), gotErr)
			},
		},
		"upload dir but building image fails": {
			Namespace: "some-namespace",
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{{
					ContainerRegistry: "some-reg",
					Spec:              []kf.CommandSpec{{Name: "some-name", BuildTemplate: "some-template", UploadDir: true}}}},
				}, nil
			}),
			BuildImage: func(t *testing.T) apps.SrcImageBuilder {
				return apps.SrcImageBuilderFunc(func(dir, tag string) error {
					return errors.New("some-error")
				})
			},
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertNil(t, "error", err)
				gotErr := results["some-name"].Execute()
				testutil.AssertErrorsEqual(t, errors.New("failed to upload directory: some-error"), gotErr)
			},
		},
		"upload dir with proper BuildImage config": {
			Namespace: "some-namespace",
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{{
					ContainerRegistry: "some-reg",
					Spec:              []kf.CommandSpec{{Name: "some-name", BuildTemplate: "some-template", UploadDir: true}}}},
				}, nil
			}),
			BuildImage: func(t *testing.T) apps.SrcImageBuilder {
				return apps.SrcImageBuilderFunc(func(dir, tag string) error {
					testutil.AssertEqual(t, "abs path", true, filepath.IsAbs(dir))
					testutil.AssertRegexp(t, "container name", `^some-reg\/src-[-0-9]+$`, tag)
					return nil
				})
			},
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertNil(t, "error", err)
				gotErr := results["some-name"].Execute()
				testutil.AssertNil(t, "error", gotErr)
			},
		},
		"upload dir with build envs": {
			Namespace: "some-namespace",
			ListReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &kf.CommandSetList{Items: []kf.CommandSet{{
					ContainerRegistry: "some-reg",
					Spec:              []kf.CommandSpec{{Name: "some-name", BuildTemplate: "some-template", UploadDir: true}}}},
				}, nil
			}),
			CreateBuildReactor: ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				b := action.(ktesting.CreateActionImpl).Object.(*build.Build)

				testutil.AssertRegexp(t, "STDOUT_PREFIX", `^.+$`, b.Spec.Template.Env[1].Value)
				testutil.AssertRegexp(t, "STDERR_PREFIX", `^.+$`, b.Spec.Template.Env[2].Value)
				testutil.AssertRegexp(t, "SOURCE_IMAGE", `^some-reg\/src-[-0-9]+$`, b.Spec.Template.Env[3].Value)

				// We don't need to double assert on the values
				b.Spec.Template.Env[1].Value = "" // STDOUT_PREFIX
				b.Spec.Template.Env[2].Value = "" // STDERR_PREFIX
				b.Spec.Template.Env[3].Value = "" // SOURCE_IMAGE

				testutil.AssertEqual(t, "env", []corev1.EnvVar{
					{
						Name:  "NAMESPACE",
						Value: "some-namespace",
					},
					{
						Name: "STDOUT_PREFIX",
					},
					{
						Name: "STDERR_PREFIX",
					},
					{
						Name: "SOURCE_IMAGE",
					},
				}, b.Spec.Template.Env)

				return false, nil, nil
			}),
			Assert: func(t *testing.T, results map[string]*cobra.Command, err error) {
				testutil.AssertNil(t, "error", err)
				gotErr := results["some-name"].Execute()
				testutil.AssertNil(t, "error", gotErr)
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.ListReactor == nil {
				tc.ListReactor = ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &kf.CommandSetList{Items: []kf.CommandSet{
						{},
					}}, nil
				})
			}

			if tc.CreateBuildReactor == nil {
				tc.CreateBuildReactor = ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return false, nil, nil
				})
			}

			if tc.DeleteBuildReactor == nil {
				tc.DeleteBuildReactor = ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return false, nil, nil
				})
			}

			if tc.BuildTailer == nil {
				tc.BuildTailer = func(*testing.T) pkf.BuildTailer {
					return pkf.BuildTailerFunc(func(ctx context.Context, w io.Writer, name, namespace string) error {
						return nil
					})
				}
			}

			if tc.BuildImage == nil {
				tc.BuildImage = func(*testing.T) apps.SrcImageBuilder {
					return apps.SrcImageBuilderFunc(func(string, string) error {
						return nil
					})
				}
			}

			if tc.Assert == nil {
				tc.Assert = func(*testing.T, map[string]*cobra.Command, error) {}
			}

			kfFake := &fake.FakeKfV1alpha1{
				Fake: &ktesting.Fake{},
			}
			kfFake.AddReactor("list", "*", tc.ListReactor)

			fakeBuild := &bfake.FakeBuildV1alpha1{
				Fake: &ktesting.Fake{},
			}

			fakeBuild.AddReactor("create", "*", tc.CreateBuildReactor)
			fakeBuild.AddReactor("delete", "*", tc.DeleteBuildReactor)

			f := utils.NewCommandOverrideFetcher(kfFake, fakeBuild, tc.BuildTailer(t), tc.BuildImage(t), &config.KfParams{Namespace: tc.Namespace})
			results, err := f.FetchCommandOverrides()
			tc.Assert(t, results, err)
		})
	}
}
