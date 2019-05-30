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

package buildpacks_test

//go:generate mockgen --package=buildpacks_test --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_image_test.go --mock_names=Image=FakeImage,Layer=FakeLayer github.com/google/go-containerregistry/pkg/v1 Image,Layer

import (
	"archive/tar"
	"bytes"
	"errors"
	io "io"
	"io/ioutil"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/testutil"
	"github.com/buildpack/pack"
	gomock "github.com/golang/mock/gomock"
	"github.com/google/go-containerregistry/pkg/name"
	gcrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestClient_UploadBuildTemplate(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		ImageName            string
		ExpectedErr          error
		BuildTemplateErr     error
		BuildTemplateItems   []string
		ListBuildTemplateErr error
		HandleDeployAction   func(t *testing.T, action ktesting.Action)
		HandleListAction     func(t *testing.T, action ktesting.Action)
	}{
		"sets meta data for list": {
			ImageName: "some-image",
			HandleListAction: func(t *testing.T, action ktesting.Action) {
				testutil.AssertEqual(t, "Verb", "list", action.GetVerb())
				testutil.AssertEqual(t, "Resource", "clusterbuildtemplates", action.GetResource().Resource)
				testutil.AssertEqual(t, "FieldSelector Field", "metadata.name", action.(ktesting.ListActionImpl).ListRestrictions.Fields.Requirements()[0].Field)
				testutil.AssertEqual(t, "FieldSelector Value", "buildpack", action.(ktesting.ListActionImpl).ListRestrictions.Fields.Requirements()[0].Value)
			},
		},
		"sets meta data for create": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				testutil.AssertEqual(t, "Verb", "create", action.GetVerb())
				testutil.AssertEqual(t, "Resource", "clusterbuildtemplates", action.GetResource().Resource)

				bt := action.(ktesting.CreateActionImpl).Object.(*build.ClusterBuildTemplate)
				testutil.AssertEqual(t, "apiVersion", "build.knative.dev/v1alpha1", bt.APIVersion)
				testutil.AssertEqual(t, "kind", "ClusterBuildTemplate", bt.Kind)
				testutil.AssertEqual(t, "Name", "buildpack", bt.Name)
			},
		},
		"sets meta data for update": {
			ImageName:          "some-image",
			BuildTemplateItems: []string{"template-1"},
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				testutil.AssertEqual(t, "Verb", "update", action.GetVerb())
				testutil.AssertEqual(t, "Resource", "clusterbuildtemplates", action.GetResource().Resource)

				bt := action.(ktesting.UpdateActionImpl).Object.(*build.ClusterBuildTemplate)
				testutil.AssertEqual(t, "Resource", "template-1-version", bt.ResourceVersion)
			},
		},
		"sets the step parameters": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				bt := action.(ktesting.CreateActionImpl).Object.(*build.ClusterBuildTemplate)
				params := map[interface{}]interface{}{}
				for _, p := range bt.Spec.Parameters {
					if p.Default == nil {
						params[p.Name] = ""
						continue
					}

					params[p.Name] = *p.Default
				}

				testutil.AssertKeyWithValue(t, params, "IMAGE", "")
				testutil.AssertKeyWithValue(t, params, "RUN_IMAGE", "packs/run:v3alpha2")
				testutil.AssertKeyWithValue(t, params, "BUILDER_IMAGE", "some-image")
				testutil.AssertKeyWithValue(t, params, "USE_CRED_HELPERS", "true")
				testutil.AssertKeyWithValue(t, params, "CACHE", "empty-dir")
				testutil.AssertKeyWithValue(t, params, "USER_ID", "1000")
				testutil.AssertKeyWithValue(t, params, "GROUP_ID", "1000")
				testutil.AssertKeyWithValue(t, params, "BUILDPACK", "")
			},
		},
		"step prepare": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				step := action.(ktesting.CreateActionImpl).Object.(*build.ClusterBuildTemplate).Spec.Steps[0]
				testutil.AssertEqual(t, "Name", "prepare", step.Name)
				testutil.AssertEqual(t, "Image", "alpine", step.Image)
				testutil.AssertEqual(t, "Command", []string{"/bin/sh"}, step.Command)
				testutil.AssertEqual(t, "Args", []string{
					"-c",
					`chown -R "${USER_ID}:${GROUP_ID}" "/builder/home" &&
						 chown -R "${USER_ID}:${GROUP_ID}" /layers &&
						 chown -R "${USER_ID}:${GROUP_ID}" /workspace`,
				}, step.Args)
				testutil.AssertEqual(t, "VolumeMounts.Name", "${CACHE}", step.VolumeMounts[0].Name)
				testutil.AssertEqual(t, "VolumeMounts.MountPath", "/layers", step.VolumeMounts[0].MountPath)
				testutil.AssertEqual(t, "ImagePullPolicy", "Always", string(step.ImagePullPolicy))
			},
		},
		"step detect": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				step := action.(ktesting.CreateActionImpl).Object.(*build.ClusterBuildTemplate).Spec.Steps[1]
				testutil.AssertEqual(t, "Name", "detect", step.Name)
				testutil.AssertEqual(t, "Image", "${BUILDER_IMAGE}", step.Image)
				testutil.AssertEqual(t, "Command", []string{"/bin/bash"}, step.Command)
				testutil.AssertEqual(t, "VolumeMounts.Name", "${CACHE}", step.VolumeMounts[0].Name)
				testutil.AssertEqual(t, "VolumeMounts.MountPath", "/layers", step.VolumeMounts[0].MountPath)
				testutil.AssertEqual(t, "ImagePullPolicy", "Always", string(step.ImagePullPolicy))
			},
		},
		"step analyze": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				step := action.(ktesting.CreateActionImpl).Object.(*build.ClusterBuildTemplate).Spec.Steps[2]
				testutil.AssertEqual(t, "Name", "analyze", step.Name)
				testutil.AssertEqual(t, "Image", "${BUILDER_IMAGE}", step.Image)
				testutil.AssertEqual(t, "Command", []string{"/lifecycle/analyzer"}, step.Command)
				testutil.AssertEqual(t, "Args", []string{
					"-layers=/layers",
					"-helpers=${USE_CRED_HELPERS}",
					"-group=/layers/group.toml",
					"${IMAGE}",
				}, step.Args)
				testutil.AssertEqual(t, "VolumeMounts.Name", "${CACHE}", step.VolumeMounts[0].Name)
				testutil.AssertEqual(t, "VolumeMounts.MountPath", "/layers", step.VolumeMounts[0].MountPath)
				testutil.AssertEqual(t, "ImagePullPolicy", "Always", string(step.ImagePullPolicy))
			},
		},
		"step build": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				step := action.(ktesting.CreateActionImpl).Object.(*build.ClusterBuildTemplate).Spec.Steps[3]
				testutil.AssertEqual(t, "Name", "build", step.Name)
				testutil.AssertEqual(t, "Image", "${BUILDER_IMAGE}", step.Image)
				testutil.AssertEqual(t, "Command", []string{"/lifecycle/builder"}, step.Command)
				testutil.AssertEqual(t, "Args", []string{
					"-layers=/layers",
					"-app=/workspace",
					"-group=/layers/group.toml",
					"-plan=/layers/plan.toml",
				}, step.Args)
				testutil.AssertEqual(t, "VolumeMounts.Name", "${CACHE}", step.VolumeMounts[0].Name)
				testutil.AssertEqual(t, "VolumeMounts.MountPath", "/layers", step.VolumeMounts[0].MountPath)
				testutil.AssertEqual(t, "ImagePullPolicy", "Always", string(step.ImagePullPolicy))
			},
		},
		"step export": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				step := action.(ktesting.CreateActionImpl).Object.(*build.ClusterBuildTemplate).Spec.Steps[4]
				testutil.AssertEqual(t, "Name", "export", step.Name)
				testutil.AssertEqual(t, "Image", "${BUILDER_IMAGE}", step.Image)
				testutil.AssertEqual(t, "Command", []string{"/lifecycle/exporter"}, step.Command)
				testutil.AssertEqual(t, "Args", []string{
					"-layers=/layers",
					"-helpers=${USE_CRED_HELPERS}",
					"-app=/workspace",
					"-image=${RUN_IMAGE}",
					"-group=/layers/group.toml",
					"${IMAGE}",
				}, step.Args)
				testutil.AssertEqual(t, "VolumeMounts.Name", "${CACHE}", step.VolumeMounts[0].Name)
				testutil.AssertEqual(t, "VolumeMounts.MountPath", "/layers", step.VolumeMounts[0].MountPath)
				testutil.AssertEqual(t, "ImagePullPolicy", "Always", string(step.ImagePullPolicy))
			},
		},
		"volumes": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				volumes := action.(ktesting.CreateActionImpl).Object.(*build.ClusterBuildTemplate).Spec.Volumes
				testutil.AssertEqual(t, "Volumes.Name", "empty-dir", volumes[0].Name)
			},
		},
		"empty image name returns an error": {
			ImageName:   "",
			ExpectedErr: errors.New("image name must not be empty"),
		},
		"upload build template error": {
			ImageName:        "some-image",
			BuildTemplateErr: errors.New("some-error"),
			ExpectedErr:      errors.New("some-error"),
		},
		"list build templates error": {
			ImageName:            "some-image",
			ListBuildTemplateErr: errors.New("some-error"),
			ExpectedErr:          errors.New("some-error"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			fake := &fake.FakeBuildV1alpha1{
				Fake: &ktesting.Fake{},
			}
			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if action.GetVerb() == "list" {
					if tc.HandleListAction != nil {
						tc.HandleListAction(t, action)
					}
					if tc.ListBuildTemplateErr != nil {
						return true, nil, tc.ListBuildTemplateErr
					}

					if len(tc.BuildTemplateItems) > 0 {
						var bt build.ClusterBuildTemplateList
						for _, name := range tc.BuildTemplateItems {
							bt.Items = append(bt.Items, build.ClusterBuildTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Name:            name,
									ResourceVersion: name + "-version",
								},
							})
						}

						return true, &bt, nil
					}
				}

				if action.GetVerb() == "create" || action.GetVerb() == "update" {
					if tc.HandleDeployAction != nil {
						tc.HandleDeployAction(t, action)
					}
					return tc.BuildTemplateErr != nil, nil, tc.BuildTemplateErr
				}

				return false, nil, nil
			}))

			c := buildpacks.NewClient(
				fake, // cbuild.BuildV1alpha1Interface
				nil,  // RemoteImageFetcher
				nil,  // BuilderFactoryCreate
			)
			gotErr := c.UploadBuildTemplate(tc.ImageName)
			if gotErr != nil || tc.ExpectedErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, gotErr)
				return
			}
		})
	}
}

func TestClient_List(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		ReactorListErr         error
		ExpectedErr            error
		EmptyBuildTemplateList bool
		HandleListAction       func(t *testing.T, action ktesting.Action)
		RemoteImageFetcher     func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error)
		HandleOutput           func(t *testing.T, buildpacks []string)
	}{
		"list only buildpack build template": {
			HandleListAction: func(t *testing.T, action ktesting.Action) {
				testutil.AssertEqual(t, "Verb", "list", action.GetVerb())
				testutil.AssertEqual(t, "Resource", "clusterbuildtemplates", action.GetResource().Resource)
				testutil.AssertEqual(t, "FieldSelector Field", "metadata.name", action.(ktesting.ListActionImpl).ListRestrictions.Fields.Requirements()[0].Field)
				testutil.AssertEqual(t, "FieldSelector Value", "buildpack", action.(ktesting.ListActionImpl).ListRestrictions.Fields.Requirements()[0].Value)
			},
		},
		"handles empty list of build templates": {
			EmptyBuildTemplateList: true,
			ExpectedErr:            nil,
		},
		"fetch image with default keychain": {
			RemoteImageFetcher: func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				testutil.AssertEqual(t, "image name", "index.docker.io/library/some-image:latest", ref.Name())
				fakeImage := NewFakeImage(gomock.NewController(t))
				fakeImage.EXPECT().Layers().AnyTimes()
				return fakeImage, nil
			},
		},
		"fetching container layers fails": {
			ExpectedErr: errors.New("some-error"),
			RemoteImageFetcher: func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				fakeImage := NewFakeImage(gomock.NewController(t))
				fakeImage.EXPECT().Layers().Return(nil, errors.New("some-error"))
				return fakeImage, nil
			},
		},
		"uncompressing layer fails": {
			ExpectedErr: errors.New("some-error"),
			RemoteImageFetcher: func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				fakeLayer := NewFakeLayer(gomock.NewController(t))
				fakeLayer.EXPECT().Uncompressed().Return(nil, errors.New("some-error"))

				fakeImage := NewFakeImage(gomock.NewController(t))
				fakeImage.EXPECT().Layers().Return([]gcrv1.Layer{fakeLayer}, nil)
				return fakeImage, nil
			},
		},
		"reads buldpack from order.toml in container": {
			RemoteImageFetcher: func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				testutil.AssertEqual(t, "image name", "index.docker.io/library/some-image:latest", ref.Name())
				fakeLayer := NewFakeLayer(gomock.NewController(t))
				fakeLayer.EXPECT().Uncompressed().Return(buildTar(t), nil)

				fakeImage := NewFakeImage(gomock.NewController(t))
				fakeImage.EXPECT().Layers().Return([]gcrv1.Layer{fakeLayer}, nil)
				return fakeImage, nil
			},
			HandleOutput: func(t *testing.T, buildpacks []string) {
				testutil.AssertEqual(t, "buildpack", "io.buildpacks.samples.nodejs", buildpacks[0])
				testutil.AssertEqual(t, "buildpack", "io.buildpacks.samples.go", buildpacks[1])
				testutil.AssertEqual(t, "buildpack", "io.buildpacks.samples.java", buildpacks[2])
			},
		},
		"fetching image fails": {
			ExpectedErr: errors.New("some-error"),
			RemoteImageFetcher: func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				return nil, errors.New("some-error")
			},
		},
		"fetching image layers fails": {
			ExpectedErr: errors.New("some-error"),
			RemoteImageFetcher: func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				fakeImage := NewFakeImage(gomock.NewController(t))
				fakeImage.EXPECT().Layers().Return(nil, errors.New("some-error"))
				return fakeImage, nil
			},
		},
		"empty number of layers": {
			ExpectedErr: nil, // should not fail
			RemoteImageFetcher: func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				fakeImage := NewFakeImage(gomock.NewController(t))
				fakeImage.EXPECT().Layers().Return(nil, nil)
				return fakeImage, nil
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			fake := &fake.FakeBuildV1alpha1{
				Fake: &ktesting.Fake{},
			}

			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if tc.HandleListAction != nil {
					tc.HandleListAction(t, action)
				}

				return true, buildTemplateList(tc.EmptyBuildTemplateList), tc.ReactorListErr
			}))

			rif := func(ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				if tc.RemoteImageFetcher == nil {
					fakeImage := NewFakeImage(gomock.NewController(t))
					fakeImage.EXPECT().Layers().AnyTimes()
					return fakeImage, nil
				}

				return tc.RemoteImageFetcher(t, ref, options...)
			}

			c := buildpacks.NewClient(
				fake, // cbuild.BuildV1alpha1Interface
				rif,  // RemoteImageFetcher
				nil,  // BuilderFactoryCreate
			)

			bps, gotErr := c.List()
			if gotErr != nil || tc.ExpectedErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, gotErr)
				return
			}

			if tc.HandleOutput != nil {
				tc.HandleOutput(t, bps)
			}
		})
	}
}

func buildTemplateList(empty bool) *build.ClusterBuildTemplateList {
	if empty {
		return &build.ClusterBuildTemplateList{}
	}
	builderImageDefault := "some-image"
	return &build.ClusterBuildTemplateList{
		Items: []build.ClusterBuildTemplate{{
			Spec: build.BuildTemplateSpec{
				Parameters: []build.ParameterSpec{
					{
						Name: "some-param",
					},
					{
						Name:    "BUILDER_IMAGE",
						Default: &builderImageDefault,
					},
				},
			},
		}},
	}
}

func buildTar(t *testing.T) io.ReadCloser {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	var files = []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"/buildpacks/order.toml", `
[[groups]]

  [[groups.buildpacks]]
    id = "io.buildpacks.samples.nodejs"
    version = "latest"

[[groups]]

  [[groups.buildpacks]]
    id = "io.buildpacks.samples.go"
    version = "latest"

[[groups]]

  [[groups.buildpacks]]
    id = "io.buildpacks.samples.java"
    version = "latest"
`},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	return ioutil.NopCloser(&buf)
}

func TestClient_Create(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		Dir               string
		ContainerRegistry string
		ExpectedErr       error
		Creator           buildpacks.BuilderFactoryCreate
	}{
		"empty dir": {
			Dir:               "",
			ContainerRegistry: "some-reg.io",
			ExpectedErr:       errors.New("dir must not be empty"),
		},
		"empty container registry": {
			Dir:               "some-path/builder.toml",
			ContainerRegistry: "",
			ExpectedErr:       errors.New("containerRegistry must not be empty"),
		},
		"returns an error if creating fails": {
			Dir:               "some-path/builder.toml",
			ContainerRegistry: "some-registry.io",
			Creator:           func(f pack.CreateBuilderFlags) error { return errors.New("some-error") },
			ExpectedErr:       errors.New("some-error"),
		},
		"sets the flags up": {
			Dir:               "some-path/builder.toml",
			ContainerRegistry: "some-registry.io",
			Creator: func(f pack.CreateBuilderFlags) error {
				testutil.AssertEqual(t, "publish", true, f.Publish)
				testutil.AssertEqual(t, "BuilderTomlPath", "some-path/builder.toml", f.BuilderTomlPath)
				testutil.AssertEqual(t, "RepoName", "some-path/builder.toml", f.BuilderTomlPath)
				testutil.AssertRegexp(t, "RepoName", `some-registry.io/buildpack-builder:[0-9]+`, f.RepoName)
				return nil
			},
		},
		"appends builder.toml if necessary": {
			Dir:               "some-path",
			ContainerRegistry: "some-registry.io",
			Creator: func(f pack.CreateBuilderFlags) error {
				testutil.AssertEqual(t, "BuilderTomlPath", "some-path/builder.toml", f.BuilderTomlPath)
				return nil
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.Creator == nil {
				tc.Creator = func(f pack.CreateBuilderFlags) error { return nil }
			}

			c := buildpacks.NewClient(
				nil,        // cbuild.BuildV1alpha1Interface
				nil,        // RemoteImageFetcher
				tc.Creator, // BuilderFactoryCreate
			)
			image, gotErr := c.Create(tc.Dir, tc.ContainerRegistry)
			if gotErr != nil || tc.ExpectedErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, gotErr)
				return
			}

			testutil.AssertRegexp(t, "image name", tc.ContainerRegistry+`/buildpack-builder:[0-9]+`, image)
		})
	}
}
