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

package commands

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/spaces/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestTargetCommand(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		setup  func(*testing.T, *config.KfParams, *fake.FakeClient)
		assert func(*testing.T, string, error)
		args   []string
	}{
		{
			name: "too many args",
			args: []string{"too", "many"},
			assert: func(t *testing.T, output string, err error) {
				testutil.AssertErrorsEqual(t, errors.New("accepts at most 1 arg(s), received 2"), err)
			},
		},
		{
			name: "no args - print space",
			args: nil,
			setup: func(t *testing.T, p *config.KfParams, fakeSpaces *fake.FakeClient) {
				p.Space = "some-space"
			},
			assert: func(t *testing.T, output string, err error) {
				testutil.AssertErrorsEqual(t, nil, err)
				testutil.AssertEqual(t, "output", "some-space\n", output)
			},
		},
		{
			name: "args and flags",
			args: []string{"-s=foo", "bar"},
			assert: func(t *testing.T, output string, err error) {
				testutil.AssertErrorsEqual(t, errors.New("--space (or --target-space) can't be used when the Space is provided via arguments."), err)
			},
		},
		{
			name: "targets space which does not exist",
			args: []string{"-s=missing"},
			setup: func(t *testing.T, p *config.KfParams, fakeSpaces *fake.FakeClient) {
				err := apierrors.NewNotFound(v1alpha1.Resource("spaces"), "missing")
				fakeSpaces.EXPECT().Get(gomock.Any(), "missing").Return(nil, err)
			},
			assert: func(t *testing.T, output string, err error) {
				testutil.AssertErrorsEqual(t, errors.New("Space \"missing\" doesn't exist"), err)
			},
		},
		{
			name: "tries to update space and fails",
			args: []string{"-s=foo"},
			setup: func(t *testing.T, p *config.KfParams, fakeSpaces *fake.FakeClient) {
				fakeSpaces.EXPECT().Get(gomock.Any(), "foo").Return(&v1alpha1.Space{}, nil)
				p.Config = "/<invalid filename>/<doesn't exist>/"
			},
			assert: func(t *testing.T, output string, err error) {
				testutil.AssertErrorsEqual(t, errors.New("open /<invalid filename>/<doesn't exist>/: no such file or directory"), err)
			},
		},
		{
			name: "updates Space via flag",
			args: []string{"-s=foo"},
			setup: func(t *testing.T, p *config.KfParams, fakeSpaces *fake.FakeClient) {
				fakeSpaces.EXPECT().Get(gomock.Any(), "foo").Return(&v1alpha1.Space{}, nil)
				tempDir, err := ioutil.TempDir("", "")
				testutil.AssertErrorsEqual(t, nil, err)
				configPath := filepath.Join(tempDir, "config")
				p.Config = configPath
				// p.Config = "/<invalid filename>/<doesn't exist>/"
				t.Cleanup(func() {
					defer os.RemoveAll(tempDir)
					if t.Failed() {
						return
					}

					pp, err := config.Load(configPath, p)
					testutil.AssertErrorsEqual(t, nil, err)
					testutil.AssertEqual(t, "space", "foo", pp.Space)
				})
			},
			assert: func(t *testing.T, output string, err error) {
				testutil.AssertErrorsEqual(t, nil, err)
			},
		},
		{
			name: "updates Space via arg",
			args: []string{"foo"},
			setup: func(t *testing.T, p *config.KfParams, fakeSpaces *fake.FakeClient) {
				fakeSpaces.EXPECT().Get(gomock.Any(), "foo").Return(&v1alpha1.Space{}, nil)
				tempDir, err := ioutil.TempDir("", "")
				testutil.AssertErrorsEqual(t, nil, err)
				configPath := filepath.Join(tempDir, "config")
				p.Config = configPath
				// p.Config = "/<invalid filename>/<doesn't exist>/"
				t.Cleanup(func() {
					defer os.RemoveAll(tempDir)
					if t.Failed() {
						return
					}

					pp, err := config.Load(configPath, p)
					testutil.AssertErrorsEqual(t, nil, err)
					testutil.AssertEqual(t, "space", "foo", pp.Space)
				})
			},
			assert: func(t *testing.T, output string, err error) {
				testutil.AssertErrorsEqual(t, nil, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &config.KfParams{}

			buf := bytes.Buffer{}
			ctrl := gomock.NewController(t)

			fakeSpaces := fake.NewFakeClient(ctrl)
			cmd := NewTargetCommand(p, fakeSpaces)

			// XXX: We can't simply use the NewTargetCommand directly because
			// we have to access its parent command as well for the Space
			// flag.
			fakeParent := &cobra.Command{
				Use:                   "fake",
				TraverseChildrenHooks: true,
			}
			fakeParent.PersistentFlags().String("space", "", "")
			fakeParent.AddCommand(cmd)
			fakeParent.PersistentPreRun = func(*cobra.Command, []string) {
				// We have to simulate what the root command does where it
				// reads the data from a config.
				if tc.setup != nil {
					tc.setup(t, p, fakeSpaces)
				}
			}

			fakeParent.SetArgs(append([]string{"target"}, tc.args...))
			fakeParent.SetOut(&buf)
			err := fakeParent.Execute()

			if tc.assert != nil {
				tc.assert(t, buf.String(), err)
			}
		})
	}
}
