// Copyright 2022 Google LLC
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

package servicebindings

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	appsfake "github.com/google/kf/v2/pkg/kf/apps/fake"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	serviceinstancebindingsfake "github.com/google/kf/v2/pkg/kf/serviceinstancebindings/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"knative.dev/pkg/kmeta"
)

func TestFixOrphanedBindings(t *testing.T) {
	t.Parallel()

	appA := v1alpha1.App{}
	appA.UID = "aaaa-aaaa-aaaa-aaaa"
	appA.Name = "app-a"

	bindingA := v1alpha1.ServiceInstanceBinding{}
	bindingA.Name = "binding-a"
	bindingA.ObjectMeta.OwnerReferences = append(bindingA.ObjectMeta.OwnerReferences, *kmeta.NewControllerRef(&appA))
	bindingA.Spec.App = &v1alpha1.AppRef{Name: "app-a"}

	type fakes struct {
		servicebindings *serviceinstancebindingsfake.FakeClient
		apps            *appsfake.FakeClient
	}
	cases := map[string]struct {
		space       string
		args        []string
		setup       func(*testing.T, fakes)
		wantErr     error
		wantOutputs []string
	}{
		"wrong number of args": {
			args:    []string{"example.com"},
			wantErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"listing bindings fails": {
			space: "some-namespace",
			setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			wantErr: errors.New("failed to list Bindings: some-error"),
		},
		"listing apps fails": {
			space: "some-namespace",
			setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().List(gomock.Any(), gomock.Any())
				fakes.apps.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errors.New("some-error"))
			},
			wantErr: errors.New("failed to list Apps: some-error"),
		},
		"empty space": {
			space: "some-namespace",
			setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.ServiceInstanceBinding{}, nil)
				fakes.apps.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.App{}, nil)
			},
			wantOutputs: []string{"Processed 0 binding(s), 0 owners fixed, 0 deleted"},
		},
		"valid ref": {
			space: "some-namespace",
			setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.ServiceInstanceBinding{
					bindingA,
				}, nil)
				fakes.apps.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.App{
					appA,
				}, nil)
			},
			wantOutputs: []string{"Processed 1 binding(s), 0 owners fixed, 0 deleted"},
		},
		"deleted app": {
			space: "some-namespace",
			setup: func(t *testing.T, fakes fakes) {
				missingApp := v1alpha1.ServiceInstanceBinding{}
				missingApp.Spec.App = &v1alpha1.AppRef{Name: "missing"}
				fakes.servicebindings.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.ServiceInstanceBinding{
					missingApp,
				}, nil)

				fakes.apps.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.App{}, nil)
			},
		},
		"dry run by default": {
			space: "some-namespace",
			setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.ServiceInstanceBinding{}, nil)
				fakes.apps.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.App{}, nil)
			},
			wantOutputs: []string{"Run with --dry-run=false to apply"},
		},
		"dry-run delete": {
			space: "some-namespace",
			args:  []string{"--dry-run=true"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.ServiceInstanceBinding{
					bindingA,
				}, nil)
				fakes.apps.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.App{}, nil)
			},
			wantOutputs: []string{
				"Processed 1 binding(s), 0 owners fixed, 1 deleted",
				"Run with --dry-run=false to apply",
			},
		},
		"dry-run false actuates deletes": {
			space: "some-namespace",
			args:  []string{"--dry-run=false"},
			setup: func(t *testing.T, fakes fakes) {
				fakes.servicebindings.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.ServiceInstanceBinding{
					bindingA,
				}, nil)
				fakes.apps.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.App{}, nil)

				fakes.servicebindings.EXPECT().Delete(gomock.Any(), gomock.Eq("some-namespace"), gomock.Eq("binding-a"))
			},
			wantOutputs: []string{"Processed 1 binding(s), 0 owners fixed, 1 deleted"},
		},
		"dry-run transforms": {
			space: "some-namespace",
			args:  []string{"--dry-run=true"},
			setup: func(t *testing.T, fakes fakes) {
				bindingANoOwner := v1alpha1.ServiceInstanceBinding{}
				bindingANoOwner.Spec.App = &v1alpha1.AppRef{Name: appA.Name}

				fakes.servicebindings.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.ServiceInstanceBinding{
					bindingANoOwner,
				}, nil)
				fakes.apps.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.App{
					appA,
				}, nil)
			},
			wantOutputs: []string{
				"Processed 1 binding(s), 1 owners fixed, 0 deleted",
				"Run with --dry-run=false to apply",
			},
		},
		"dry-run false actuates transforms": {
			space: "some-namespace",
			args:  []string{"--dry-run=false"},
			setup: func(t *testing.T, fakes fakes) {
				bindingANoOwner := v1alpha1.ServiceInstanceBinding{}
				bindingANoOwner.Spec.App = &v1alpha1.AppRef{Name: appA.Name}

				fakes.servicebindings.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.ServiceInstanceBinding{
					bindingANoOwner,
				}, nil)
				fakes.apps.EXPECT().List(gomock.Any(), gomock.Any()).Return([]v1alpha1.App{
					appA,
				}, nil)

				fakes.servicebindings.EXPECT().Transform(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			wantOutputs: []string{"Processed 1 binding(s), 1 owners fixed, 0 deleted"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			sbClient := serviceinstancebindingsfake.NewFakeClient(ctrl)
			appsClient := appsfake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakes{
					servicebindings: sbClient,
					apps:            appsClient,
				})
			}

			var buffer bytes.Buffer
			cmd := NewFixOrphanedBindingsCommand(
				&config.KfParams{
					Space: tc.space,
				},
				sbClient,
				appsClient,
			)

			if tc.args == nil {
				// We have to set this to something that is non-nil so that
				// cobra doesn't go use os.Args (which might have test flags
				// set).
				tc.args = make([]string, 0)
			}

			cmd.SetArgs(tc.args)
			cmd.SetOutput(&buffer)

			gotErr := cmd.Execute()
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
			testutil.AssertContainsAll(t, buffer.String(), tc.wantOutputs)
		})
	}
}
