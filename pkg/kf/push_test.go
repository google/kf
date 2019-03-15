package kf_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	kffake "github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	"github.com/golang/mock/gomock"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestPush_BadConfig(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName           string
		containerRegistry string
		serviceAccount    string
		wantErr           error
	}{
		"empty app name, returns error": {
			wantErr:           errors.New("invalid app name"),
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
		},
		"container registry not configured, returns error": {
			wantErr:        errors.New("container registry is not set"),
			serviceAccount: "some-service-account",
			appName:        "some-app",
		},
		"service account not configured, returns error": {
			wantErr:           errors.New("service account is not set"),
			containerRegistry: "some-reg.io",
			appName:           "some-app",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			p := kf.NewPusher(
				nil, // AppLister - Should not be used
				nil, // ServingFactory - Should not be used
				nil, // SrcImageBuilder - Should not be used
				nil, // Logs - Should not be used
			)

			var opts []kf.PushOption
			if tc.containerRegistry != "" {
				opts = append(opts, kf.WithPushContainerRegistry(tc.containerRegistry))
			}
			if tc.serviceAccount != "" {
				opts = append(opts, kf.WithPushServiceAccount(tc.serviceAccount))
			}

			gotErr := p.Push(tc.appName, opts...)
			if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
				t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
			}
		})
	}
}

func TestPush_Logs(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName string
		wantErr error
		logErr  error
	}{
		"fetching logs succeeds": {
			appName: "some-app",
		},
		"fetching logs returns an error, no error": {
			appName: "some-app",
			wantErr: errors.New("some error"),
			logErr:  errors.New("some error"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			expectedNamespace := "some-namespace"

			fakeAppLister := kffake.NewFakeLister(ctrl)
			fakeAppLister.
				EXPECT().
				List(gomock.Any()).
				AnyTimes()

			fakeLogs := kffake.NewFakeLogTailer(ctrl)
			fakeLogs.
				EXPECT().
				Tail(
					gomock.Not(gomock.Nil()), // out,
					tc.appName+"-version",    // resourceVersion
					expectedNamespace,        // namespace
				).
				Return(tc.logErr)

			fakeServing := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}
			fakeServing.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				// Set the ResourceVersion
				obj := action.(ktesting.CreateAction).GetObject()
				obj.(*serving.Service).ResourceVersion = tc.appName + "-version"

				return true, obj, nil
			}))

			p := kf.NewPusher(
				fakeAppLister,
				func() (cserving.ServingV1alpha1Interface, error) {
					return fakeServing, nil
				},
				func(dir, tag string) error {
					return nil
				},
				fakeLogs,
			)

			gotErr := p.Push(
				tc.appName,
				kf.WithPushNamespace(expectedNamespace),
				kf.WithPushContainerRegistry("some-container-registry"),
				kf.WithPushServiceAccount("some-service-account"),
			)

			if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
				t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
			}

			ctrl.Finish()
		})
	}
}

func TestPush_UpdateApp(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName   string
		listerErr error
		wantErr   error
	}{
		"app already exists, update": {
			appName: "some-app",
		},
		"service list error, returns error": {
			wantErr:   errors.New("some error"),
			listerErr: errors.New("some error"),
			appName:   "some-app",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			expectedNamespace := "some-namespace"
			deployedApps := []string{"some-other-app", "some-app"}
			containerRegistry := "some-reg.io"
			serviceAccount := "some-service-account"

			fakeAppLister := kffake.NewFakeLister(ctrl)
			fakeAppLister.
				EXPECT().
				List(gomock.Any()).
				DoAndReturn(func(opts ...kf.ListOption) ([]serving.Service, error) {
					if namespace := kf.ListOptions(opts).Namespace(); namespace != expectedNamespace {
						t.Fatalf("expected namespace %s, got %s", expectedNamespace, namespace)
					}

					var apps []serving.Service
					for _, appName := range deployedApps {
						s := serving.Service{}
						s.Name = appName
						s.ResourceVersion = appName + "-version"
						apps = append(apps, s)
					}

					return apps, tc.listerErr
				})

			fakeLogs := kffake.NewFakeLogTailer(ctrl)
			fakeLogs.
				EXPECT().
				Tail(gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes()

			fakeServing := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}
			var reactorCalled bool
			fakeServing.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				reactorCalled = true
				testPushReaction(t, action, expectedNamespace, tc.appName, containerRegistry, serviceAccount, "update")

				return false, nil, nil
			}))

			p := kf.NewPusher(
				fakeAppLister,
				func() (cserving.ServingV1alpha1Interface, error) {
					return fakeServing, nil
				},
				func(dir, tag string) error {
					return nil
				},
				fakeLogs,
			)

			gotErr := p.Push(
				tc.appName,
				kf.WithPushNamespace(expectedNamespace),
				kf.WithPushContainerRegistry(containerRegistry),
				kf.WithPushServiceAccount(serviceAccount),
			)
			if tc.wantErr != nil || gotErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				return
			}

			if !reactorCalled {
				t.Fatal("Reactor was not invoked")
			}

			ctrl.Finish()
		})
	}
}

func TestPush_NewApp(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		namespace         string
		appName           string
		path              string
		wantErr           error
		servingFactoryErr error
		serviceCreateErr  error
	}{
		"pushes app to a configured namespace": {
			namespace: "some-namespace",
			appName:   "some-app",
		},
		"pushes app to default namespace": {
			appName: "some-app",
		},
		"serving factory error, returns error": {
			wantErr:           errors.New("some error"),
			servingFactoryErr: errors.New("some error"),
			appName:           "some-app",
		},
		"service create error, returns error": {
			wantErr:          errors.New("some error"),
			serviceCreateErr: errors.New("some error"),
			appName:          "some-app",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			containerRegistry := "some-reg.io"
			serviceAccount := "some-service-account"

			fake := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}

			expectedNamespace := tc.namespace
			if tc.namespace == "" {
				expectedNamespace = "default"
			}
			expectedPath := tc.path
			if tc.path == "" {
				cwd, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}
				expectedPath = cwd
			}

			var reactorCalled bool
			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				reactorCalled = true
				testPushReaction(t, action, expectedNamespace, tc.appName, containerRegistry, serviceAccount, "create")

				return tc.serviceCreateErr != nil, nil, tc.serviceCreateErr
			}))

			fakeAppLister := kffake.NewFakeLister(ctrl)
			fakeAppLister.
				EXPECT().
				List(gomock.Any()).
				AnyTimes()

			fakeLogs := kffake.NewFakeLogTailer(ctrl)
			fakeLogs.
				EXPECT().
				Tail(gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes()

			srcBuilder := func(dir, tag string) error {
				if dir != expectedPath {
					t.Fatalf("wanted image dir %s, got %s", expectedPath, dir)
				}

				if !strings.HasPrefix(tag, containerRegistry) {
					t.Fatalf("want container registry prefix %s, got %s", containerRegistry, tag)
				}

				if !strings.HasSuffix(tag, "latest") {
					t.Fatalf("want container registry suffix %s, got %s", "latest", tag)
				}
				return nil
			}

			p := kf.NewPusher(
				fakeAppLister,
				func() (cserving.ServingV1alpha1Interface, error) {
					return fake, tc.servingFactoryErr
				},
				srcBuilder,
				fakeLogs,
			)

			opts := []kf.PushOption{
				kf.WithPushContainerRegistry(containerRegistry),
				kf.WithPushServiceAccount(serviceAccount),
			}
			if tc.namespace != "" {
				opts = append(opts, kf.WithPushNamespace(tc.namespace))
			}
			if tc.path != "" {
				opts = append(opts, kf.WithPushPath(tc.path))
			}

			gotErr := p.Push(tc.appName, opts...)
			if tc.wantErr != nil || gotErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				return
			}

			if !reactorCalled {
				t.Fatal("Reactor was not invoked")
			}

			ctrl.Finish()
		})
	}
}

func testPushReaction(
	t *testing.T,
	action ktesting.Action,
	namespace string,
	appName string,
	containerRegistry string,
	serviceAccount string,
	actionVerb string,
) {
	t.Helper()

	if action.GetNamespace() != namespace {
		t.Fatalf("wanted namespace: %s, got: %s", namespace, action.GetNamespace())
	}

	if !action.Matches(actionVerb, "services") {
		t.Fatal("wrong action")
	}

	service := action.(ktesting.CreateAction).GetObject().(*serving.Service)
	imageName := testBuild(t, appName, containerRegistry, serviceAccount, service.Spec.RunLatest.Configuration.Build)
	testRevisionTemplate(t, imageName, service.Spec.RunLatest.Configuration.RevisionTemplate)

	if service.Name != appName {
		t.Errorf("wanted service name %s, got %s", appName, service.Name)
	}

	if service.Kind != "Service" {
		t.Errorf("wanted service Kind %s, got %s", "Service", service.Kind)
	}

	if service.APIVersion != "serving.knative.dev/v1alpha1" {
		t.Errorf("wanted service APIVersion %s, got %s", "serving.knative.dev/v1alpha1", service.APIVersion)
	}

	if service.Namespace != namespace {
		t.Errorf("wanted service Namespace %s, got %s", namespace, service.Namespace)
	}

	if actionVerb == "update" && service.ResourceVersion != appName+"-version" {
		t.Errorf("wanted service ResourceVersion (on update) %s, got %s", appName+"-version", service.ResourceVersion)
	}
}

func testRevisionTemplate(t *testing.T, imageName string, spec serving.RevisionTemplateSpec) {
	t.Helper()

	if spec.Spec.Container.Image != imageName {
		t.Errorf("wanted image name %s, got %s", imageName, spec.Spec.Container.Image)
	}

	if spec.Spec.Container.ImagePullPolicy != "Always" {
		t.Errorf("wanted image pull policy %s, got %s", "Always", spec.Spec.Container.ImagePullPolicy)
	}
}

func testBuild(
	t *testing.T,
	appName string,
	containerRegistry string,
	serviceAccount string,
	raw *serving.RawExtension,
) string {
	t.Helper()

	var b build.Build
	if err := json.Unmarshal(raw.Raw, &b); err != nil {
		t.Fatal(err)
	}

	if b.Spec.ServiceAccountName != serviceAccount {
		t.Errorf("wanted service account name: %s, got %s", serviceAccount, b.Spec.ServiceAccountName)
	}

	srcPattern := fmt.Sprintf(`^%s/src-%s-[0-9]{19}:latest$`, containerRegistry, appName)
	if !regexp.MustCompile(srcPattern).MatchString(b.Spec.Source.Custom.Image) {
		t.Errorf("wanted image pattern: %s, got %s", srcPattern, b.Spec.Source.Custom.Image)
	}

	if b.Spec.Template.Name != "buildpack" {
		t.Errorf("wanted template name: %s, got %s", "buildpack", b.Spec.Template.Name)
	}

	if len(b.Spec.Template.Arguments) != 1 {
		t.Fatalf("wanted template args len: 1, got %d", len(b.Spec.Template.Arguments))
	}

	if b.Spec.Template.Arguments[0].Name != "IMAGE" {
		t.Errorf("wanted template args name: %s, got %s", "IMAGE", b.Spec.Template.Arguments[0].Name)
	}

	imageName := b.Spec.Template.Arguments[0].Value
	prefix := fmt.Sprintf("%s/%s", containerRegistry, appName)
	if !strings.HasPrefix(imageName, prefix) {
		t.Errorf("wanted image name to have prefix %s, got: %s", prefix, imageName)
	}

	pattern := fmt.Sprintf(`^%s/%s-[0-9]{19}:latest$`, containerRegistry, appName)
	if !regexp.MustCompile(pattern).MatchString(imageName) {
		t.Errorf("wanted image name pattern: %s, got %s", pattern, imageName)
	}

	return imageName
}
