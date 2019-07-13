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

package quotas

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf/commands/config"
	"github.com/google/kf/pkg/kf/spaces"
	"github.com/google/kf/pkg/kf/spaces/fake"
	"github.com/google/kf/pkg/kf/testutil"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestUpdateQuotaCommand(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		namespace string
		quotaName string
		wantErr   error
		args      []string
		setup     func(t *testing.T, fakeUpdater *fake.FakeClient)
		assert    func(t *testing.T, buffer *bytes.Buffer)
	}{
		"invalid number of args": {
			args:    []string{},
			wantErr: errors.New("accepts 1 arg(s), received 0"),
		},
		"update error": {
			args:    []string{"some-space", "-m", "100z"},
			wantErr: errors.New("some-error"),
			setup: func(t *testing.T, fakeUpdater *fake.FakeClient) {
				fakeUpdater.
					EXPECT().
					Transform(gomock.Any(), gomock.Any()).
					Return(errors.New("some-error"))
			},
		},
		"some flags": {
			args:      []string{"some-quota", "-m", "1024M"},
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeUpdater *fake.FakeClient) {
				fakeUpdater.
					EXPECT().
					Transform(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		"update success": {
			args:      []string{"some-quota", "-m", "20Gi"},
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeUpdater *fake.FakeClient) {
				fakeUpdater.
					EXPECT().
					Transform(gomock.Any(), gomock.Any()).
					Do(func(spaceName string, transformer spaces.Mutator) error {
						kfspace, err := newDummyKfSpace("1024M", "4")
						testutil.AssertNil(t, "Parse resource quantity err", err)
						transformer(kfspace.ToSpace())

						expectedMemory, memErr := resource.ParseQuantity("20Gi")
						testutil.AssertNil(t, "Parse memory quantity err", memErr)
						expectedCPU, cpuErr := resource.ParseQuantity("4")
						testutil.AssertNil(t, "Parse cpu quantity err", cpuErr)

						actualMemory, _ := kfspace.GetMemory()
						actualCPU, _ := kfspace.GetCPU()
						testutil.AssertEqual(t, "Updated memory", expectedMemory, actualMemory)
						testutil.AssertEqual(t, "Same CPU", expectedCPU, actualCPU)
						return err
					})
			},
		},
		"reset quota success": {
			args:      []string{"some-quota", "-m", "0"},
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeUpdater *fake.FakeClient) {
				fakeUpdater.
					EXPECT().
					Transform(gomock.Any(), gomock.Any()).
					Do(func(spaceName string, transformer spaces.Mutator) error {
						kfspace, err := newDummyKfSpace("1024M", "4")
						testutil.AssertNil(t, "Parse resource quantity err", err)
						transformer(kfspace.ToSpace())

						expectedCPU, cpuErr := resource.ParseQuantity("4")
						testutil.AssertNil(t, "Parse cpu quantity err", cpuErr)

						_, quotaExists := kfspace.GetMemory()
						actualCPU, _ := kfspace.GetCPU()
						testutil.AssertEqual(t, "Memory quota exists", false, quotaExists)
						testutil.AssertEqual(t, "Same CPU", expectedCPU, actualCPU)
						return err
					})
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeUpdater := fake.NewFakeClient(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeUpdater)
			}

			buffer := &bytes.Buffer{}

			c := NewUpdateQuotaCommand(&config.KfParams{
				Namespace: tc.namespace,
			}, fakeUpdater)
			c.SetOutput(buffer)

			c.SetArgs(tc.args)
			gotErr := c.Execute()
			if tc.wantErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			if tc.assert != nil {
				tc.assert(t, buffer)
			}

			testutil.AssertNil(t, "Command err", gotErr)
			testutil.AssertEqual(t, "SilenceUsage", true, c.SilenceUsage)

			ctrl.Finish()
		})
	}
}

func newDummyKfSpace(memory string, cpu string) (*spaces.KfSpace, error) {
	kfspace := spaces.NewKfSpace()
	memQuantity, memErr := resource.ParseQuantity(memory)
	if memErr != nil {
		return &kfspace, memErr
	}
	cpuQuantity, cpuErr := resource.ParseQuantity(cpu)
	if cpuErr != nil {
		return &kfspace, cpuErr
	}

	kfspace.SetMemory(memQuantity)
	kfspace.SetCPU(cpuQuantity)
	return &kfspace, nil
}
