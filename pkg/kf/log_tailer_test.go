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
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/pkg/kf"
	"github.com/google/kf/pkg/kf/testutil"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/pkg/apis"
	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servicefake "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

//go:generate mockgen --package kf_test --destination fake_watcher_test.go --mock_names=Interface=FakeWatcher --copyright_file internal/tools/option-builder/LICENSE_HEADER k8s.io/apimachinery/pkg/watch Interface

func TestLogTailer_DeployLogs_ServiceLogs(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName                  string
		namespace                string
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
			events:          append(createMsgEvents("some-app", "", "msg-1", "msg-2"), createMsgEvents("some-app", "True", "")...),
			wantedMsgs:      []string{"msg-1", "msg-2"},
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
			events:          createMsgEvents("some-app", "False", "some-error"),
			wantErr:         errors.New("deployment failed: some-error"),
		},
		"watch fails, return error": {
			appName:         "some-app",
			namespace:       "default",
			resourceVersion: "some-version",
			events:          nil,
			wantErr:         errors.New("lost connection to Kubernetes"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl, fakeServing := buildLogWatchFakes(
				t,
				tc.events, nil,
				tc.serviceWatchErr, nil,
			)

			fakeServing.PrependWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				testWatch(t, action, "services", tc.namespace, tc.resourceVersion)
				return false, nil, nil
			}))

			lt := kf.NewLogTailer(fakeServing)

			var buffer bytes.Buffer
			gotErr := lt.DeployLogs(&buffer, tc.appName, tc.resourceVersion, tc.namespace)
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

func createMsgEvents(appName string, status corev1.ConditionStatus, msgs ...string) []watch.Event {
	var es []watch.Event
	for _, m := range msgs {
		es = append(es, watch.Event{
			Object: &serving.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: appName,
				},
				Status: serving.ServiceStatus{
					Status: duckv1beta1.Status{
						Conditions: []apis.Condition{
							{Type: "Ready", Status: status, Message: m},
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
) (*gomock.Controller, *servicefake.FakeServingV1alpha1) {
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

	return ctrl, fakeServing
}
