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

import (
	"archive/tar"
	"bytes"
	"errors"
	io "io"
	"io/ioutil"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/golang/mock/gomock"
	"github.com/google/go-containerregistry/pkg/name"
	gcrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	"github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

//go:generate mockgen --package=buildpacks_test --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_image_test.go --mock_names=Image=FakeImage,Layer=FakeLayer github.com/google/go-containerregistry/pkg/v1 Image,Layer

func TestBuildpackLister(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		BuildFactoryErr        error
		ReactorListErr         error
		ExpectedErr            error
		EmptyBuildTemplateList bool
		HandleListAction       func(t *testing.T, action ktesting.Action)
		RemoteImageFetcher     func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error)
		HandleOutput           func(t *testing.T, buildpacks []string)
	}{
		"build factory returns an error": {
			BuildFactoryErr: errors.New("some-error"),
			ExpectedErr:     errors.New("some-error"),
		},
		"list only buildpack build template": {
			HandleListAction: func(t *testing.T, action ktesting.Action) {
				testutil.AssertEqual(t, "Verb", "list", action.GetVerb())
				testutil.AssertEqual(t, "Resource", "buildtemplates", action.GetResource().Resource)
				testutil.AssertEqual(t, "namespace", "default", action.GetNamespace())
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

			l := buildpacks.NewBuildpackLister(func() (cbuild.BuildV1alpha1Interface, error) {
				return fake, tc.BuildFactoryErr
			}, func(ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				if tc.RemoteImageFetcher == nil {
					fakeImage := NewFakeImage(gomock.NewController(t))
					fakeImage.EXPECT().Layers().AnyTimes()
					return fakeImage, nil
				}

				return tc.RemoteImageFetcher(t, ref, options...)
			})

			bps, gotErr := l.List()
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

func buildTemplateList(empty bool) *build.BuildTemplateList {
	if empty {
		return &build.BuildTemplateList{}
	}
	builderImageDefault := "some-image"
	return &build.BuildTemplateList{
		Items: []build.BuildTemplate{{
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
