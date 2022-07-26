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

package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"knative.dev/pkg/apis"
)

func TestBuildpackV2List_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"happy path": {
			Context: context.Background(),
			Input: BuildpackV2List{
				{Name: "some-buildpack", URL: "https://some-buildpack"},
			},
			Want: nil,
		},
		"duplicate name": {
			Context: context.Background(),
			Input: BuildpackV2List{
				{Name: "some-buildpack", URL: "https://some-buildpack"},
				{Name: "some-buildpack", URL: "https://some-buildpack"},
			},
			Want: &apis.FieldError{
				Message: "duplicate name",
				Details: `the name "some-buildpack" is duplicated`,
				Paths:   []string{"[1].name"},
			},
		},
		"recurses to children": {
			Context: context.Background(),
			Input: BuildpackV2List{
				{Name: "some-buildpack"},
			},
			Want: apis.ErrMissingField("[0].url"),
		},
	}

	cases.Run(t)
}

func TestBuildpackV2_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"happy path": {
			Context: context.Background(),
			Input: &BuildpackV2Definition{
				Name: "some-buildpack",
				URL:  "https://some-buildpack",
			},
			Want: nil,
		},
		"missing URL": {
			Context: context.Background(),
			Input: &BuildpackV2Definition{
				Name: "some-buildpack",
			},
			Want: apis.ErrMissingField("url"),
		},
		"missing name": {
			Context: context.Background(),
			Input: &BuildpackV2Definition{
				URL: "https://some-buildpack",
			},
			Want: apis.ErrMissingField("name"),
		},
	}

	cases.Run(t)
}

func TestStackV2List_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"happy path": {
			Context: context.Background(),
			Input: StackV2List{
				{Name: "some-stack", Image: "cloudfoundry/cflinuxfs3"},
			},
			Want: nil,
		},
		"duplicate name": {
			Context: context.Background(),
			Input: StackV2List{
				{Name: "some-stack", Image: "cloudfoundry/cflinuxfs3"},
				{Name: "some-stack", Image: "cloudfoundry/cflinuxfs3"},
			},
			Want: &apis.FieldError{
				Message: "duplicate name",
				Details: `the name "some-stack" is duplicated`,
				Paths:   []string{"[1].name"},
			},
		},
		"recurses to children": {
			Context: context.Background(),
			Input: StackV2List{
				{Name: "some-stack"},
			},
			Want: apis.ErrMissingField("[0].image"),
		},
	}

	cases.Run(t)
}

func TestStackV2_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"happy path": {
			Context: context.Background(),
			Input: &StackV2Definition{
				Name:  "some-stack",
				Image: "cloudfoundry/cflinuxfs3",
			},
			Want: nil,
		},
		"missing URL": {
			Context: context.Background(),
			Input: &StackV2Definition{
				Name: "some-stack",
			},
			Want: apis.ErrMissingField("image"),
		},
		"missing name": {
			Context: context.Background(),
			Input: &StackV2Definition{
				Image: "cloudfoundry/cflinuxfs3",
			},
			Want: apis.ErrMissingField("name"),
		},
	}

	cases.Run(t)
}

func TestStackV3List_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"happy path": {
			Context: context.Background(),
			Input: StackV3List{
				{Name: "some-stack", RunImage: "run", BuildImage: "build"},
			},
			Want: nil,
		},
		"duplicate name": {
			Context: context.Background(),
			Input: StackV3List{
				{Name: "some-stack", RunImage: "run", BuildImage: "build"},
				{Name: "some-stack", RunImage: "run", BuildImage: "build"},
			},
			Want: &apis.FieldError{
				Message: "duplicate name",
				Details: `the name "some-stack" is duplicated`,
				Paths:   []string{"[1].name"},
			},
		},
		"recurses to children": {
			Context: context.Background(),
			Input: StackV3List{
				{Name: "some-stack", RunImage: "cloudfoundry/cflinuxfs3:run"},
			},
			Want: apis.ErrMissingField("[0].buildImage"),
		},
	}

	cases.Run(t)
}

func TestStackV3_Validate(t *testing.T) {
	cases := testutil.ApisValidatableTestSuite{
		"happy path": {
			Context: context.Background(),
			Input: &StackV3Definition{
				Name:       "some-stack",
				BuildImage: "cloudfoundry/cflinuxfs3:build",
				RunImage:   "cloudfoundry/cflinuxfs3:run",
			},
			Want: nil,
		},
		"missing build image": {
			Context: context.Background(),
			Input: &StackV3Definition{
				Name:     "some-stack",
				RunImage: "cloudfoundry/cflinuxfs3:run",
			},
			Want: apis.ErrMissingField("buildImage"),
		},
		"missing run image": {
			Context: context.Background(),
			Input: &StackV3Definition{
				Name:       "some-stack",
				BuildImage: "cloudfoundry/cflinuxfs3:build",
			},
			Want: apis.ErrMissingField("runImage"),
		},
		"missing name": {
			Context: context.Background(),
			Input: &StackV3Definition{
				BuildImage: "cloudfoundry/cflinuxfs3:build",
				RunImage:   "cloudfoundry/cflinuxfs3:run",
			},
			Want: apis.ErrMissingField("name"),
		},
	}

	cases.Run(t)
}

func ExampleStackV2List_FindStackByName() {
	list := StackV2List{
		{
			Name:  "cflinuxfs3",
			Image: "cloudfoundry/cflinuxfs3",
		},
	}

	fmt.Println("doesn't exist:", list.FindStackByName("does-not-exist"))
	fmt.Println("exists image:", list.FindStackByName("cflinuxfs3").Image)

	// Output: doesn't exist: <nil>
	// exists image: cloudfoundry/cflinuxfs3
}

func ExampleBuildpackV2List_WithoutDisabled() {
	list := BuildpackV2List{
		{
			Name:     "java_buildpack",
			Disabled: true,
		},
		{
			Name:     "dotnet_buildpack",
			Disabled: false,
		},
		{
			Name:     "ruby_buildpack",
			Disabled: false,
		},
	}

	for _, b := range list.WithoutDisabled() {
		fmt.Println(b.Name)
	}

	// Output: dotnet_buildpack
	// ruby_buildpack
}

func ExampleBuildpackV2List_MapToURL() {
	list := BuildpackV2List{
		{
			Name: "java_buildpack",
			URL:  "https://path-to-java-buildpack",
		},
	}

	fmt.Println("custom not mapped:", list.MapToURL("https://some-custom-buildpack"))
	fmt.Println("java_buildpack:", list.MapToURL("java_buildpack"))

	// Output: custom not mapped: https://some-custom-buildpack
	// java_buildpack: https://path-to-java-buildpack
}
