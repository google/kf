package kf_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/golang/mock/gomock"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	buildfake "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1/fake"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	servicefake "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

//go:generate mockgen --package kf_test --destination fake_watcher_test.go --mock_names=Interface=FakeWatcher k8s.io/apimachinery/pkg/watch Interface

func TestLogTailer(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name              string
		namespace         string
		resourceVersion   string
		added             bool
		buildFactoryErr   error
		servingFactoryErr error
		wantErr           error
		buildWatchErr     error
		serviceWatchErr   error
		buildTailErr      error
	}{
		{
			name:            "fetch logs for build",
			namespace:       "default",
			resourceVersion: "some-version",
			added:           true,
		},
		{
			name:            "when build is not added, don't display logs",
			namespace:       "default",
			resourceVersion: "some-version",
			added:           false,
		},
		{
			name:            "build factory returns error, return error",
			namespace:       "default",
			resourceVersion: "some-version",
			buildFactoryErr: errors.New("some-error"),
			wantErr:         errors.New("some-error"),
		},
		{
			name:            "watch build returns an error, return error",
			namespace:       "default",
			resourceVersion: "some-version",
			buildWatchErr:   errors.New("some-error"),
			wantErr:         errors.New("some-error"),
		},
		{
			name:              "serving factory returns error, return error",
			namespace:         "default",
			resourceVersion:   "some-version",
			servingFactoryErr: errors.New("some-error"),
			wantErr:           errors.New("some-error"),
		},
		{
			name:            "watch service returns an error, return error",
			namespace:       "default",
			resourceVersion: "some-version",
			serviceWatchErr: errors.New("some-error"),
			wantErr:         errors.New("some-error"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeBuildWatcher := NewFakeWatcher(ctrl)
			fakeServiceWatcher := NewFakeWatcher(ctrl)

			buildEvents := make(chan watch.Event, 1)
			eventType := watch.Modified
			if tc.added {
				eventType = watch.Added
			}
			b := &build.Build{}
			b.Name = "build-name"
			buildEvents <- watch.Event{
				Type:   eventType,
				Object: b,
			}
			close(buildEvents)

			fakeBuildWatcher.
				EXPECT().
				ResultChan().
				DoAndReturn(func() <-chan watch.Event {
					return buildEvents
				})

			// Ensure Stop is invoked to clean up resources.
			fakeBuildWatcher.
				EXPECT().
				Stop()

			msgs := []string{"msg-a", "msg-b"}
			fakeServiceWatcher.
				EXPECT().
				ResultChan().
				DoAndReturn(func() <-chan watch.Event {
					c := make(chan watch.Event, len(msgs))
					for _, m := range msgs {
						c <- watch.Event{
							Object: &serving.Service{
								Status: serving.ServiceStatus{
									Conditions: duckv1alpha1.Conditions{
										{Message: m},
									},
								},
							},
						}
					}
					close(c)
					return c
				})

			// Ensure Stop is invoked to clean up resources.
			fakeServiceWatcher.
				EXPECT().
				Stop()

			fakeBuild := &buildfake.FakeBuildV1alpha1{
				Fake: &ktesting.Fake{},
			}

			var (
				buildReactorCalled   bool
				buildLogCalled       bool
				servingReactorCalled bool
				buffer               bytes.Buffer
			)

			fakeBuild.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				buildReactorCalled = true

				testWatch(t, action, "builds", tc.namespace, tc.resourceVersion)

				return true, fakeBuildWatcher, tc.buildWatchErr
			}))

			fakeServing := &servicefake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}

			fakeServing.AddWatchReactor("*", ktesting.WatchReactionFunc(func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				servingReactorCalled = true

				testWatch(t, action, "services", tc.namespace, tc.resourceVersion)

				return true, fakeServiceWatcher, tc.serviceWatchErr
			}))

			lt := kf.NewLogTailer(
				func() (cbuild.BuildV1alpha1Interface, error) {
					return fakeBuild, tc.buildFactoryErr
				},
				func() (cserving.ServingV1alpha1Interface, error) {
					return fakeServing, tc.servingFactoryErr
				},
				func(ctx context.Context, out io.Writer, buildName, namespace string) error {
					if !tc.added {
						t.Fatal("build logs should not have been fetched")
					}
					buildLogCalled = true

					if buildName != "build-name" {
						t.Fatalf("wanted buildName: %s, got: %s", buildName, "build-name")
					}
					if namespace != tc.namespace {
						t.Fatalf("wanted namespace: %s, got: %s", tc.namespace, namespace)
					}
					if ctx == nil {
						t.Fatalf("wanted non-nil context")
					}
					if !reflect.DeepEqual(out, &buffer) {
						t.Fatalf("wrong out, wanted buffer")
					}

					return tc.buildTailErr
				},
			)

			gotErr := lt.Tail(&buffer, tc.resourceVersion, tc.namespace)
			if tc.wantErr != nil || gotErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				return
			}

			if !buildReactorCalled {
				t.Fatal("Build Reactor was not invoked")
			}
			if !servingReactorCalled {
				t.Fatal("Serving Reactor was not invoked")
			}
			if tc.added && !buildLogCalled {
				t.Fatal("BuildLog was not invoked")
			}

			for _, msg := range msgs {
				if strings.Index(buffer.String(), msg) < 0 {
					t.Fatalf("wanted %q to contain %q", buffer.String(), msg)
				}
			}

			ctrl.Finish()
		})
	}
}

func testWatch(t *testing.T, action ktesting.Action, resource, namespace, resourceVersion string) {
	t.Helper()

	if action.GetNamespace() != namespace {
		t.Fatalf("wanted namespace: %s, got: %s", namespace, action.GetNamespace())
	}

	if !action.Matches("watch", resource) {
		t.Fatal("wrong action")
	}

	if rv := action.(ktesting.WatchActionImpl).WatchRestrictions.ResourceVersion; rv != resourceVersion {
		t.Fatalf("wanted resourceVersion %s, got %s", resourceVersion, rv)
	}
}
