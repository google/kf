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

package spaces

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/spaces/fake"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
)

func TestNewGetSpaceCommand(t *testing.T) {
	t.Parallel()

	goodSpace := &v1alpha1.Space{}
	goodSpace.Name = "my-space"
	goodSpace.Spec = v1alpha1.SpaceSpec{}
	goodSpace.Spec.Security.EnableDeveloperLogsAccess = true
	goodSpace.Status.Conditions = []apis.Condition{{
		Type:   "Ready",
		Status: "TESTING",
		Reason: "SomeMessage",
	}}
	goodSpace.Spec.BuildpackBuild.BuilderImage = "some/builder/image"
	goodSpace.Spec.BuildpackBuild.ContainerRegistry = "some/container/registry"
	goodSpace.Spec.BuildpackBuild.Env = []corev1.EnvVar{{Name: "BuildVar", Value: "BuildVal"}}
	goodSpace.Spec.Execution.Env = []corev1.EnvVar{{Name: "ExecVar", Value: "ExecVal"}}

	cases := map[string]struct {
		wantErr    error
		args       []string
		space      *v1alpha1.Space
		wantOutput []string
	}{
		"invalid number of args": {
			args:    []string{},
			wantErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"metadata": {
			args:       []string{"my-space"},
			space:      goodSpace,
			wantOutput: []string{"Metadata", "my-space", "Ready", "SomeMessage"},
		},
		"security": {
			args:       []string{"my-space"},
			space:      goodSpace,
			wantOutput: []string{"Security", "read logs?", "true"},
		},
		"build": {
			args:       []string{"my-space"},
			space:      goodSpace,
			wantOutput: []string{"Build", "some/builder/image", "some/container/registry", "BuildVar", "BuildVal"},
		},
		"execution": {
			args:       []string{"my-space"},
			space:      goodSpace,
			wantOutput: []string{"Execution", "ExecVar", "ExecVal"},
		},
		"client error": {
			args:    []string{"my-space"},
			space:   nil,
			wantErr: errors.New("does not exist"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeSpaces := fake.NewFakeClient(ctrl)

			if tc.space != nil {
				fakeSpaces.EXPECT().Get(gomock.Any()).Return(tc.space, nil)
			} else {
				fakeSpaces.EXPECT().Get(gomock.Any()).Return(nil, errors.New("does not exist"))
			}

			buffer := &bytes.Buffer{}

			c := NewGetSpaceCommand(&config.KfParams{Namespace: "default"}, fakeSpaces)
			c.SetOutput(buffer)
			c.SetArgs(tc.args)

			gotErr := c.Execute()
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)

			if tc.wantErr == nil {
				testutil.AssertContainsAll(t, buffer.String(), tc.wantOutput)

				ctrl.Finish()
			}
		})
	}
}
