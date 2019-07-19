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

package apps_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	v1alpha1fake "github.com/google/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1/fake"
	"github.com/google/kf/pkg/kf/apps"
	sourcesfake "github.com/google/kf/pkg/kf/sources/fake"
	systemenvinjectorfake "github.com/google/kf/pkg/kf/systemenvinjector/fake"
	"github.com/google/kf/pkg/kf/testutil"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

//go:generate mockgen --package apps_test --destination fake_watcher_test.go --mock_names=Interface=FakeWatcher --copyright_file ../internal/tools/option-builder/LICENSE_HEADER k8s.io/apimachinery/pkg/watch Interface

func TestLogTailer_DeployLogs_ServiceLogs(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName                  string
		namespace                string
		noStart                  bool
		resourceVersion          string
		serviceWatchErr          error
		events                   []watch.Event
		unwantedMsgs, wantedMsgs []string
		wantErr                  error
	}{
		"displays deployment messages": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			events: createMsgEvents("some-app", duckv1beta1.Conditions{
				{
					Type:    "SourceReady",
					Status:  "True",
					Message: "msg-1",
				},
				{
					Type:    "Ready",
					Status:  "True",
					Message: "msg-2",
				},
			},
			),
			wantedMsgs: []string{"msg-1", "msg-2"},
		},
		"NoStart don't display deployment messages": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			noStart:         true,
			events: createMsgEvents("some-app", duckv1beta1.Conditions{
				{
					Type:    "SourceReady",
					Status:  "True",
					Message: "msg-1",
				},
				{
					Type:    "Ready",
					Status:  "True",
					Message: "msg-2",
				},
			},
			),
			wantedMsgs:   []string{"msg-1"},
			unwantedMsgs: []string{"msg-2"},
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
			events: createMsgEvents("some-app", duckv1beta1.Conditions{
				{
					Type:    "SourceReady",
					Status:  "True",
					Message: "some-error",
				},
				{
					Type:    "Ready",
					Status:  "False",
					Message: "some-error",
				},
			}),
			wantErr: errors.New("deployment failed: some-error"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl, fakeApps := buildLogWatchFakes(
				t,
				tc.events, nil,
				tc.serviceWatchErr, nil,
			)

			fakeApps.PrependWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				testWatch(t, action, "apps", tc.namespace, tc.resourceVersion)
				return false, nil, nil
			}))

			sourceClient := sourcesfake.NewFakeClient(ctrl)
			seif := systemenvinjectorfake.NewFakeSystemEnvInjector(ctrl)
			lt := apps.NewClient(fakeApps, seif, sourceClient)

			var buffer bytes.Buffer
			gotErr := lt.DeployLogs(
				&buffer,
				tc.appName,
				tc.resourceVersion,
				tc.namespace,
				tc.noStart,
			)
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

func testWatch(t *testing.T, action ktesting.Action, resource, namespace, resourceVersion string) {
	t.Helper()
	testutil.AssertEqual(t, "namespace", namespace, action.GetNamespace())
	testutil.AssertEqual(t, "resourceVersion", resourceVersion, action.(ktesting.WatchActionImpl).WatchRestrictions.ResourceVersion)

	if !action.Matches("watch", resource) {
		t.Fatalf("wrong action: %s", resource)
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

func createMsgEvents(appName string, conditions duckv1beta1.Conditions) []watch.Event {
	var es []watch.Event
	es = append(es, watch.Event{
		Object: &v1alpha1.App{
			ObjectMeta: metav1.ObjectMeta{
				Name: appName,
			},
			Status: v1alpha1.AppStatus{
				Status: duckv1beta1.Status{
					Conditions: conditions,
				},
			},
		},
	})
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
) (*gomock.Controller, *v1alpha1fake.FakeKfV1alpha1) {
	ctrl := gomock.NewController(t)
	fakeWatcher := NewFakeWatcher(ctrl)
	fakeWatcher.
		EXPECT().
		ResultChan().
		DoAndReturn(func() <-chan watch.Event {
			return createEvents(serviceEvents)
		}).
		AnyTimes()

	// Ensure Stop is invoked to clean up resources.
	fakeWatcher.
		EXPECT().
		Stop().
		AnyTimes()

	fakeKfClient := &v1alpha1fake.FakeKfV1alpha1{
		Fake: &ktesting.Fake{},
	}

	fakeKfClient.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, fakeWatcher, serviceErr
	}))

	return ctrl, fakeKfClient
}
