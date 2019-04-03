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
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

//go:generate mockgen --package kf_test --destination fake_watcher_test.go --mock_names=Interface=FakeWatcher k8s.io/apimachinery/pkg/watch Interface

func TestLogTailer_ServiceLogs(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		namespace         string
		resourceVersion   string
		servingFactoryErr error
		serviceWatchErr   error
		events            []watch.Event
		wantedMsgs        []string
		wantErr           error
	}{
		"displays deployment messages": {
			namespace:       "default",
			resourceVersion: "some-version",
			events:          createMsgEvents("", "msg-1", "msg-2"),
			wantedMsgs:      []string{"msg-1", "msg-2"},
		},
		"serving factory returns error, return error": {
			namespace:         "default",
			resourceVersion:   "some-version",
			servingFactoryErr: errors.New("some-error"),
			wantErr:           errors.New("some-error"),
		},
		"watch service returns an error, return error": {
			namespace:       "default",
			resourceVersion: "some-version",
			serviceWatchErr: errors.New("some-error"),
			wantErr:         errors.New("some-error"),
		},
		"revision fails, return error": {
			namespace:       "default",
			resourceVersion: "some-version",
			events:          createMsgEvents("RevisionFailed", "some-error"),
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
			gotErr := lt.Tail(&buffer, tc.resourceVersion, tc.namespace)
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			for _, msg := range tc.wantedMsgs {
				if strings.Index(buffer.String(), msg) < 0 {
					t.Fatalf("wanted %q to contain %q", buffer.String(), msg)
				}
			}

			ctrl.Finish()
		})
	}
}

func TestLogTailer_BuildLogs(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		namespace       string
		resourceVersion string
		buildFactoryErr error
		buildWatchErr   error
		buildTailErr    error
		events          []watch.Event
		wantedMsgs      []string
		wantErr         error
	}{
		"fetch logs for build": {
			namespace:       "default",
			resourceVersion: "some-version",
			events:          createBuildAddedEvent(),
		},
		"build factory returns error, return error": {
			namespace:       "default",
			resourceVersion: "some-version",
			buildFactoryErr: errors.New("some-error"),
			wantErr:         errors.New("some-error"),
		},
		"build tail returns error, return error": {
			namespace:       "default",
			resourceVersion: "some-version",
			buildTailErr:    errors.New("some-error"),
			events:          createBuildAddedEvent(),
			wantErr:         errors.New("some-error"),
		},
		"build fails, returns error": {
			namespace:       "default",
			resourceVersion: "some-version",
			events: append(createBuildAddedEvent(), watch.Event{
				Object: &build.Build{
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

			var buffer bytes.Buffer
			lt := kf.NewLogTailer(
				func() (cbuild.BuildV1alpha1Interface, error) {
					return fakeBuild, tc.buildFactoryErr
				},
				func() (cserving.ServingV1alpha1Interface, error) {
					return fakeServing, nil
				},
				func(ctx context.Context, out io.Writer, buildName, namespace string) error {
					testutil.AssertEqual(t, "buildName", "build-name", buildName)
					testutil.AssertEqual(t, "namespace", tc.namespace, namespace)
					testutil.AssertEqual(t, "out", &buffer, out)

					return tc.buildTailErr
				},
			)

			gotErr := lt.Tail(&buffer, tc.resourceVersion, tc.namespace)
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
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

func createMsgEvents(reason string, msgs ...string) []watch.Event {
	var es []watch.Event
	for _, m := range msgs {
		es = append(es, watch.Event{
			Object: &serving.Service{
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

func createBuildAddedEvent() []watch.Event {
	b := &build.Build{}
	b.Name = "build-name"
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
