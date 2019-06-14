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

//go:generate mockgen --package=buildpacks_test --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_image_test.go --mock_names=Image=FakeImage github.com/google/go-containerregistry/pkg/v1 Image

import (
	"errors"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/testutil"
	gomock "github.com/golang/mock/gomock"
	"github.com/google/go-containerregistry/pkg/name"
	gcrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestClient_List(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		ReactorListErr         error
		ExpectedErr            error
		EmptyBuildTemplateList bool
		HandleListAction       func(t *testing.T, action ktesting.Action)
		RemoteImageFetcher     func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error)
		HandleOutput           func(t *testing.T, buildpacks []buildpacks.Buildpack)
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
				setEmptyConfig(fakeImage)
				return fakeImage, nil
			},
		},
		"fetching container ConfigFile fails": {
			ExpectedErr: errors.New("some-error"),
			RemoteImageFetcher: func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				fakeImage := NewFakeImage(gomock.NewController(t))
				fakeImage.EXPECT().ConfigFile().Return(nil, errors.New("some-error"))
				return fakeImage, nil
			},
		},
		"unmarshalling MetaDataLabel fails": {
			ExpectedErr: errors.New("EOF"),
			RemoteImageFetcher: func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				fakeImage := NewFakeImage(gomock.NewController(t))
				fakeImage.EXPECT().ConfigFile().Return(&gcrv1.ConfigFile{
					Config: gcrv1.Config{
						Labels: nil, // Empty so it will fail to parse
					},
				}, nil)
				return fakeImage, nil
			},
		},
		"reads buldpack from label in container": {
			RemoteImageFetcher: func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				testutil.AssertEqual(t, "image name", "index.docker.io/library/some-image:latest", ref.Name())
				fakeImage := NewFakeImage(gomock.NewController(t))
				fakeImage.EXPECT().ConfigFile().Return(&gcrv1.ConfigFile{
					Config: gcrv1.Config{
						Labels: map[string]string{
							"io.buildpacks.builder.metadata": `{"buildpacks":[{"id":"io.buildpacks.samples.nodejs"},{"id":"io.buildpacks.samples.go"},{"id":"io.buildpacks.samples.java"}]}`,
						},
					},
				}, nil)
				return fakeImage, nil
			},
			HandleOutput: func(t *testing.T, buildpacks []buildpacks.Buildpack) {
				testutil.AssertEqual(t, "buildpack", "io.buildpacks.samples.nodejs", buildpacks[0].ID)
				testutil.AssertEqual(t, "buildpack", "io.buildpacks.samples.go", buildpacks[1].ID)
				testutil.AssertEqual(t, "buildpack", "io.buildpacks.samples.java", buildpacks[2].ID)
			},
		},
		"fetching image fails": {
			ExpectedErr: errors.New("some-error"),
			RemoteImageFetcher: func(t *testing.T, ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error) {
				return nil, errors.New("some-error")
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
					setEmptyConfig(fakeImage)
					return fakeImage, nil
				}

				return tc.RemoteImageFetcher(t, ref, options...)
			}

			c := buildpacks.NewClient(
				fake, // cbuild.BuildV1alpha1Interface
				rif,  // RemoteImageFetcher
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

func setEmptyConfig(fakeImage *FakeImage) {
	fakeImage.EXPECT().ConfigFile().Return(&gcrv1.ConfigFile{
		Config: gcrv1.Config{
			Labels: map[string]string{
				"io.buildpacks.builder.metadata": `{}`,
			},
		},
	}, nil).AnyTimes()
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
