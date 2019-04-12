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

package kf_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/golang/mock/gomock"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	buildfake "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1/fake"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	servicefake "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

//go:generate mockgen --package kf_test --destination fake_watcher_test.go --mock_names=Interface=FakeWatcher --copyright_file internal/tools/option-builder/LICENSE_HEADER k8s.io/apimachinery/pkg/watch Interface

func TestLogTailer_ServiceLogs(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName                  string
		namespace                string
		resourceVersion          string
		servingFactoryErr        error
		serviceWatchErr          error
		events                   []watch.Event
		unwantedMsgs, wantedMsgs []string
		wantErr                  error
	}{
		"displays deployment messages": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			events:          createMsgEvents("some-app", "", "msg-1", "msg-2"),
			wantedMsgs:      []string{"msg-1", "msg-2"},
		},
		"filters out messages from other apps": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			events:          append(createMsgEvents("some-app", "", "msg-1", "msg-2"), createMsgEvents("some-other-app", "", "should not see")...),
			wantedMsgs:      []string{"msg-1", "msg-2"},
			unwantedMsgs:    []string{"should not see"},
		},
		"serving factory returns error, return error": {
			appName:           "some-app",
			namespace:         "default",
			resourceVersion:   "some-version",
			servingFactoryErr: errors.New("some-error"),
			wantErr:           errors.New("some-error"),
		},
		"watch service returns an error, return error": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			serviceWatchErr: errors.New("some-error"),
			wantErr:         errors.New("some-error"),
		},
		"revision fails, return error": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			events:          createMsgEvents("some-app", "RevisionFailed", "some-error"),
			wantErr:         errors.New("deployment failed: some-error"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl, fakeServing, fakeBuild := buildLogWatchFakes(
				t,
				tc.events, nil,
				tc.serviceWatchErr, nil,
			)

			fakeServing.PrependWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				testWatch(t, action, "services", tc.namespace, tc.resourceVersion)
				return false, nil, nil
			}))

			lt := kf.NewLogTailer(
				func() (cbuild.BuildV1alpha1Interface, error) {
					return fakeBuild, nil
				},
				func() (cserving.ServingV1alpha1Interface, error) {
					return fakeServing, tc.servingFactoryErr
				},
				func(ctx context.Context, out io.Writer, buildName, namespace string) error {
					return nil
				},
			)

			var buffer bytes.Buffer
			gotErr := lt.Tail(&buffer, tc.appName, tc.resourceVersion, tc.namespace)
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			for _, msg := range tc.wantedMsgs {
				if strings.Index(buffer.String(), msg) < 0 {
					t.Fatalf("wanted %q to contain %q", buffer.String(), msg)
				}
			}

			for _, msg := range tc.unwantedMsgs {
				if strings.Index(buffer.String(), msg) >= 0 {
					t.Fatalf("wanted %q to not contain %q", buffer.String(), msg)
				}
			}

			ctrl.Finish()
		})
	}
}

func TestLogTailer_BuildLogs(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName         string
		namespace       string
		resourceVersion string
		buildFactoryErr error
		buildWatchErr   error
		events          []watch.Event
		wantedMsgs      []string
		wantErr         error
		bufferF         func(t *testing.T, buf *bytes.Buffer)
		tailF           func(t *testing.T) func(ctx context.Context, out io.Writer, buildName, namespace string) error
	}{
		"fetch logs for build": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			events:          append(createBuildAddedEvent("some-other-app", "some-other-build"), createBuildAddedEvent("some-app", "some-build")...),
			tailF: func(t *testing.T) func(ctx context.Context, out io.Writer, buildName, namespace string) error {
				return func(ctx context.Context, out io.Writer, buildName, namespace string) error {
					testutil.AssertEqual(t, "buildName", "some-build", buildName)
					testutil.AssertEqual(t, "namespace", "default", namespace)

					out.Write([]byte("some-data"))

					return nil
				}
			},
			bufferF: func(t *testing.T, buf *bytes.Buffer) {
				testutil.AssertContainsAll(t, buf.String(), []string{"some-data"})
			},
		},
		"build factory returns error, return error": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			buildFactoryErr: errors.New("some-error"),
			wantErr:         errors.New("some-error"),
		},
		"build tail returns error, return error": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			events:          createBuildAddedEvent("some-app", "some-build"),
			wantErr:         errors.New("some-error"),
			tailF: func(t *testing.T) func(ctx context.Context, out io.Writer, buildName, namespace string) error {
				return func(ctx context.Context, out io.Writer, buildName, namespace string) error {
					return errors.New("some-error")
				}
			},
		},
		"build fails, returns error": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			events: append(createBuildAddedEvent("some-app", "some-build"), watch.Event{
				Object: &build.Build{
					ObjectMeta: metav1.ObjectMeta{
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "some-app",
							},
						},
					},
					Status: build.BuildStatus{
						Status: duckv1alpha1.Status{
							Conditions: duckv1alpha1.Conditions{
								{
									Type:    "Succeeded",
									Status:  "False",
									Message: "some-message",
								},
							},
						},
					},
				},
			}),
			wantErr: errors.New("build failed: some-message"),
		},
		"watch build returns an error, return error": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			buildWatchErr:   errors.New("some-error"),
			wantErr:         errors.New("some-error"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl, fakeServing, fakeBuild := buildLogWatchFakes(
				t,
				nil, tc.events,
				nil, tc.buildWatchErr,
			)

			fakeBuild.PrependWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				testWatch(t, action, "builds", tc.namespace, tc.resourceVersion)
				return false, nil, nil
			}))

			if tc.tailF == nil {
				tc.tailF = func(t *testing.T) func(ctx context.Context, out io.Writer, buildName, namespace string) error {
					return func(ctx context.Context, out io.Writer, buildName, namespace string) error {
						return nil
					}
				}
			}

			var buffer bytes.Buffer
			lt := kf.NewLogTailer(
				func() (cbuild.BuildV1alpha1Interface, error) {
					return fakeBuild, tc.buildFactoryErr
				},
				func() (cserving.ServingV1alpha1Interface, error) {
					return fakeServing, nil
				},
				tc.tailF(t),
			)

			gotErr := lt.Tail(&buffer, tc.appName, tc.resourceVersion, tc.namespace)
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			if tc.bufferF != nil {
				tc.bufferF(t, &buffer)
			}

			ctrl.Finish()
		})
	}
}

func testWatch(t *testing.T, action ktesting.Action, resource, namespace, resourceVersion string) {
	t.Helper()

	testutil.AssertEqual(t, "namespace", namespace, action.GetNamespace())
	testutil.AssertEqual(t, "resourceVersion", resourceVersion, action.(ktesting.WatchActionImpl).WatchRestrictions.ResourceVersion)

	if !action.Matches("watch", resource) {
		t.Fatal("wrong action")
	}
}

func createEvents(es []watch.Event) <-chan watch.Event {
	c := make(chan watch.Event, len(es))
	defer close(c)
	for _, e := range es {
		c <- e
	}
	return c
}

func createMsgEvents(appName, reason string, msgs ...string) []watch.Event {
	var es []watch.Event
	for _, m := range msgs {
		es = append(es, watch.Event{
			Object: &serving.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: appName,
				},
				Status: serving.ServiceStatus{
					Status: duckv1alpha1.Status{
						Conditions: duckv1alpha1.Conditions{
							{Reason: reason, Message: m},
						},
					},
				},
			},
		})
	}
	return es
}

func createBuildAddedEvent(appName, buildName string) []watch.Event {
	b := &build.Build{}
	b.Name = buildName
	b.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		{
			Name: appName,
		},
	}
	return []watch.Event{
		{
			Type:   watch.Added,
			Object: b,
		},
	}
}

func buildLogWatchFakes(
	t *testing.T,
	serviceEvents, buildEvents []watch.Event,
	serviceErr, buildErr error,
) (*gomock.Controller, *servicefake.FakeServingV1alpha1, *buildfake.FakeBuildV1alpha1) {
	ctrl := gomock.NewController(t)
	fakeServiceWatcher := NewFakeWatcher(ctrl)
	fakeServiceWatcher.
		EXPECT().
		ResultChan().
		DoAndReturn(func() <-chan watch.Event {
			return createEvents(serviceEvents)
		})

	// Ensure Stop is invoked to clean up resources.
	fakeServiceWatcher.
		EXPECT().
		Stop().
		AnyTimes()

	fakeServing := &servicefake.FakeServingV1alpha1{
		Fake: &ktesting.Fake{},
	}

	fakeServing.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, fakeServiceWatcher, serviceErr
	}))

	fakeBuildWatcher := NewFakeWatcher(ctrl)

	fakeBuildWatcher.
		EXPECT().
		ResultChan().
		DoAndReturn(func() <-chan watch.Event {
			return createEvents(buildEvents)
		})

	// Ensure Stop is invoked to clean up resources.
	fakeBuildWatcher.
		EXPECT().
		Stop().
		AnyTimes()

	fakeBuild := &buildfake.FakeBuildV1alpha1{
		Fake: &ktesting.Fake{},
	}

	fakeBuild.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, fakeBuildWatcher, buildErr
	}))

	return ctrl, fakeServing, fakeBuild
}
