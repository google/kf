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

package logs_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/kf/pkg/kf/logs"
	"github.com/google/kf/pkg/kf/testutil"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/typed/core/v1/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestTailer_Tail(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		AppName string
		Opts    []logs.TailOption
		Setup   func(t *testing.T, fake *fake.FakeCoreV1)
		Assert  func(t *testing.T, buf *bytes.Buffer, err error)
	}{
		"default namespace": {
			AppName: "some-app",
			Setup: func(t *testing.T, fake *fake.FakeCoreV1) {
				fake.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
					testutil.AssertEqual(t, "namespace", "default", action.GetNamespace())
					return false, nil, nil
				}))
			},
		},
		"custom namespace": {
			AppName: "some-app",
			Opts: []logs.TailOption{
				logs.WithTailNamespace("custom-namespace"),
			},
			Setup: func(t *testing.T, fake *fake.FakeCoreV1) {
				fake.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
					testutil.AssertEqual(t, "namespace", "custom-namespace", action.GetNamespace())
					return false, nil, nil
				}))
			},
		},
		"empty app name": {
			AppName: "",
			Assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("appName is empty"), err)
			},
		},
		"negative number of lines": {
			AppName: "some-app",
			Opts: []logs.TailOption{
				logs.WithTailNumberLines(-1),
			},
			Assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("number of lines must be greater than or equal to 0"), err)
			},
		},
		"watching pods fails": {
			AppName: "some-app",
			Setup: func(t *testing.T, fake *fake.FakeCoreV1) {
				fake.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
					return true, nil, errors.New("some-error")
				}))
			},
			Assert: func(t *testing.T, buf *bytes.Buffer, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to watch pods: some-error"), err)
			},
		},
		"uses service selector": {
			AppName: "some-app",
			Setup: func(t *testing.T, fake *fake.FakeCoreV1) {
				fake.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
					labels := action.(ktesting.WatchActionImpl).WatchRestrictions.Labels
					testutil.AssertEqual(t, "labels", "serving.knative.dev/service=some-app", labels.String())

					return false, nil, nil
				}))
			},
		},
		"writes logs to the writer": {
			AppName: "some-app",
			Setup: func(t *testing.T, fake *fake.FakeCoreV1) {
				fake.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
					labels := action.(ktesting.WatchActionImpl).WatchRestrictions.Labels
					testutil.AssertEqual(t, "labels", "serving.knative.dev/service=some-app", labels.String())

					return false, nil, nil
				}))
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.Setup == nil {
				tc.Setup = func(t *testing.T, fake *fake.FakeCoreV1) {
					// NOP
				}
			}

			if tc.Assert == nil {
				tc.Assert = func(t *testing.T, buf *bytes.Buffer, err error) {
					testutil.AssertNil(t, "err", err)
				}
			}

			fakeClient := &fake.FakeCoreV1{
				Fake: &ktesting.Fake{},
			}

			tc.Setup(t, fakeClient)
			fakeClient.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				f := watch.NewFake()
				f.Stop()
				return true, f, nil
			}))

			buf := &bytes.Buffer{}
			gotErr := logs.NewTailer(fakeClient).Tail(context.Background(), tc.AppName, buf, tc.Opts...)
			tc.Assert(t, buf, gotErr)
		})
	}
}
