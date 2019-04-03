package kf_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	kffake "github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	kfi "github.com/GoogleCloudPlatform/kf/pkg/kf/internal/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/golang/mock/gomock"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestPush_BadConfig(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		appName string
		wantErr error
		opts    kf.PushOptions
	}{
		"empty app name, returns error": {
			wantErr: errors.New("invalid app name"),
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
		},
		"container registry and docker image are NOT configured, returns error": {
			appName: "some-app",
			wantErr: errors.New("container registry or docker image must be set (not both)"),
			opts: kf.PushOptions{
				kf.WithPushServiceAccount("some-service-account"),
			},
		},
		"container registry and docker image are configured, returns error": {
			appName: "some-app",
			wantErr: errors.New("container registry or docker image must be set (not both)"),
			opts: kf.PushOptions{
				kf.WithPushServiceAccount("some-service-account"),
				kf.WithPushDockerImage("some-image"),
				kf.WithPushContainerRegistry("some-reg.io"),
			},
		},
		"path and docker image are configured, returns error": {
			appName: "some-app",
			wantErr: errors.New("path flag is not valid with docker image flag"),
			opts: kf.PushOptions{
				kf.WithPushServiceAccount("some-service-account"),
				kf.WithPushDockerImage("some-image"),
				kf.WithPushPath("some-path"),
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			p := kf.NewPusher(
				nil, // AppLister - Should not be used
				nil, // ServingFactory - Should not be used
				nil, // SrcImageBuilder - Should not be used
				nil, // Logs - Should not be used
			)

			gotErr := p.Push(tc.appName, tc.opts...)
			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)

			if !kfi.ConfigError(gotErr) {
				t.Fatal("wanted error to be a ConfigError")
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
					tc.appName,               // appName
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

			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
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
		envs      map[string]string
	}{
		"app already exists, update": {
			appName: "some-app",
		},
		"preserves existing environment variables": {
			appName: "some-app",
			envs: map[string]string{
				"ENV-1": "a",
				"ENV-2": "b",
			},
		},
		"service list error, returns error": {
			appName:   "some-app",
			wantErr:   errors.New("some error"),
			listerErr: errors.New("some error"),
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
						s := buildServiceWithEnvs(appName, tc.envs)
						s.ResourceVersion = appName + "-version"
						apps = append(apps, s)
					}

					return apps, tc.listerErr
				})

			fakeLogs := kffake.NewFakeLogTailer(ctrl)
			fakeLogs.
				EXPECT().
				Tail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes()

			fakeServing := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}
			var reactorCalled bool
			fakeServing.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				reactorCalled = true
				testPushReaction(t, action, expectedNamespace, tc.appName, containerRegistry, serviceAccount, "update", "", tc.envs)

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
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
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
		appName           string
		opts              kf.PushOptions
		wantErr           error
		servingFactoryErr error
		serviceCreateErr  error
		assert            func(t *testing.T, service *serving.Service)
	}{
		"pushes app to a configured namespace": {
			appName: "some-app",
			opts: kf.PushOptions{
				kf.WithPushNamespace("some-namespace"),
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
		},
		"pushes app to default namespace": {
			appName: "some-app",
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
		},
		"pushes docker image": {
			appName: "some-app",
			opts: kf.PushOptions{
				kf.WithPushDockerImage("some-docker-image"),
				kf.WithPushServiceAccount("some-service-account"),
			},
		},
		"serving factory error, returns error": {
			appName:           "some-app",
			wantErr:           errors.New("some error"),
			servingFactoryErr: errors.New("some error"),
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
		},
		"service create error, returns error": {
			appName:          "some-app",
			wantErr:          errors.New("some error"),
			serviceCreateErr: errors.New("some error"),
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushServiceAccount("some-service-account"),
			},
		},
		"set ports to h2c for gRPC": {
			appName: "some-app",
			assert: func(t *testing.T, service *serving.Service) {
				testutil.AssertEqual(
					t,
					"container.ports",
					[]corev1.ContainerPort{{Name: "h2c", ContainerPort: 8080}},
					service.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Ports,
				)
			},
			opts: kf.PushOptions{
				kf.WithPushContainerRegistry("some-reg.io"),
				kf.WithPushGrpc(true),
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			fake := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}

			expectedNamespace := tc.opts.Namespace()
			if tc.opts.Namespace() == "" {
				expectedNamespace = "default"
			}
			expectedPath := tc.opts.Path()
			if tc.opts.Path() == "" {
				cwd, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}
				expectedPath = cwd
			}

			var reactorCalled bool
			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				reactorCalled = true
				testPushReaction(t, action, expectedNamespace, tc.appName, tc.opts.ContainerRegistry(), tc.opts.ServiceAccount(), "create", tc.opts.DockerImage(), nil)

				if tc.assert != nil {
					tc.assert(
						t,
						action.(ktesting.CreateAction).GetObject().(*serving.Service),
					)
				}

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
				Tail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes()

			var srcBuilderCalled bool
			srcBuilder := func(dir, tag string) error {
				srcBuilderCalled = true
				if tc.opts.DockerImage() != "" {
					t.Fatal("should not have been called with docker image")
				}

				testutil.AssertEqual(t, "path", expectedPath, dir)
				testutil.AssertRegexp(t, "container registry", "^"+tc.opts.ContainerRegistry()+`/[a-zA-Z0-9_-]+:[0-9_-]+$`, tag)

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

			gotErr := p.Push(tc.appName, tc.opts...)
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			if !reactorCalled {
				t.Fatal("Reactor was not invoked")
			}
			if !srcBuilderCalled && tc.opts.DockerImage() == "" {
				t.Fatal("SrcBuilder was not invoked")
			}

			ctrl.Finish()
		})
	}
}

func TestPush_EnvVars(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		opts                       kf.PushOptions
		update                     bool
		existingEnvs, expectedEnvs map[string]string
		wantErr                    error
	}{
		"multiple envs": {
			opts: kf.PushOptions{
				kf.WithPushEnvironmentVariables([]string{
					"NAME1=VAL1",
					"NAME2=VAL2",
				}),
			},
			expectedEnvs: map[string]string{
				"NAME1": "VAL1",
				"NAME2": "VAL2",
			},
		},
		"replace envs": {
			update: true,
			opts: kf.PushOptions{
				kf.WithPushEnvironmentVariables([]string{
					"NAME1=VAL1",
					"NAME2=VAL2",
				}),
			},
			existingEnvs: map[string]string{
				"OLD1":  "VAL1",
				"OLD2":  "VAL2",
				"NAME1": "OLD1",
			},
			expectedEnvs: map[string]string{
				"NAME1": "VAL1",
				"NAME2": "VAL2",
			},
		},
		"leave existing envs": {
			update: true,
			existingEnvs: map[string]string{
				"OLD1": "VAL1",
				"OLD2": "VAL2",
			},
			expectedEnvs: map[string]string{
				"OLD1": "VAL1",
				"OLD2": "VAL2",
			},
		},
		"invalid env, returns error": {
			wantErr: errors.New("malformed environment variable: INVALID"),
			opts: kf.PushOptions{
				kf.WithPushEnvironmentVariables([]string{
					"NAME1=VAL1",
					"INVALID",
				}),
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			const appName = "some-app"
			ctrl := gomock.NewController(t)
			fakeLogs := kffake.NewFakeLogTailer(ctrl)
			fakeLogs.
				EXPECT().
				Tail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes()

			fakeAppLister := kffake.NewFakeLister(ctrl)
			var apps []serving.Service
			if tc.update {
				apps = append(apps, buildServiceWithEnvs(appName, tc.existingEnvs))
			}
			fakeAppLister.
				EXPECT().
				List(gomock.Any()).
				Return(apps, nil).
				AnyTimes()

			fake := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}

			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				service := action.(ktesting.CreateAction).GetObject().(*serving.Service)
				envs := map[string]string{}
				for _, env := range service.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env {
					envs[env.Name] = env.Value
				}

				testutil.AssertEqual(t, "env len", len(tc.expectedEnvs), len(envs))
				testutil.AssertEqual(t, "envs", tc.expectedEnvs, envs)

				return false, nil, nil
			}))

			p := kf.NewPusher(
				fakeAppLister,
				func() (cserving.ServingV1alpha1Interface, error) {
					return fake, nil
				},
				func(dir, tag string) error { return nil },
				fakeLogs,
			)

			tc.opts = append(tc.opts, kf.WithPushContainerRegistry("some-reg.io"))

			gotErr := p.Push("some-app", tc.opts...)
			if tc.wantErr != nil || gotErr != nil {
				testutil.AssertEqual(t, "ConfigErr", true, kfi.ConfigError(gotErr))
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}
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
	imageName string,
	envs map[string]string,
) {
	t.Helper()

	testutil.AssertEqual(t, "namespace", namespace, action.GetNamespace())

	if !action.Matches(actionVerb, "services") {
		t.Fatal("wrong action")
	}

	service := action.(ktesting.CreateAction).GetObject().(*serving.Service)
	if imageName == "" {
		imageName = testBuild(t, appName, containerRegistry, serviceAccount, service.Spec.RunLatest.Configuration.Build)
	} else {
		// No build
		if service.Spec.RunLatest.Configuration.Build != nil {
			t.Fatal("expected build to be nil when an image is provided")
		}
	}
	testRevisionTemplate(t, imageName, serviceAccount, service.Spec.RunLatest.Configuration.RevisionTemplate)

	testutil.AssertEqual(t, "service.Name", appName, service.Name)
	testutil.AssertEqual(t, "service.Kind", "Service", service.Kind)
	testutil.AssertEqual(t, "service.APIVersion", "serving.knative.dev/v1alpha1", service.APIVersion)
	testutil.AssertEqual(t, "service.Namespace", namespace, service.Namespace)

	if actionVerb == "update" && service.ResourceVersion != appName+"-version" {
		testutil.AssertEqual(t, "service.ResourceVersion (on update)", appName+"-version", service.ResourceVersion)
	}

	actualEnvs := service.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env
	testutil.AssertEqual(t, "env len", len(envs), len(actualEnvs))
	for _, env := range actualEnvs {
		testutil.AssertEqual(t, "env: "+env.Name, envs[env.Name], env.Value)
	}
}

func testRevisionTemplate(t *testing.T, imageName, serviceAccount string, spec serving.RevisionTemplateSpec) {
	t.Helper()

	testutil.AssertEqual(t, "Spec.Container.Image", imageName, spec.Spec.Container.Image)
	testutil.AssertEqual(t, "Spec.Container.PullPolicy", "Always", string(spec.Spec.Container.ImagePullPolicy))
	testutil.AssertEqual(t, "Spec.ServiceAccountName", serviceAccount, spec.Spec.ServiceAccountName)
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

	testutil.AssertEqual(t, "Spec.ServiceAccountName", serviceAccount, b.Spec.ServiceAccountName)

	srcPattern := fmt.Sprintf(`^%s/src-%s:[0-9]{19}$`, containerRegistry, appName)
	testutil.AssertRegexp(t, "image", srcPattern, b.Spec.Source.Custom.Image)

	testutil.AssertEqual(t, "Spec.Template.Name", "buildpack", b.Spec.Template.Name)

	if len(b.Spec.Template.Arguments) != 1 {
		t.Fatalf("wanted template args len: 1, got %d", len(b.Spec.Template.Arguments))
	}
	testutil.AssertEqual(t, "Spec.Template.Arguments[0].Name", "IMAGE", b.Spec.Template.Arguments[0].Name)

	imageName := b.Spec.Template.Arguments[0].Value

	pattern := fmt.Sprintf(`^%s/%s:[0-9]{19}$`, containerRegistry, appName)
	testutil.AssertRegexp(t, "image name", pattern, imageName)

	return imageName
}
