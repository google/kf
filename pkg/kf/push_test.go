package kf_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func TestPush(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		namespace         string
		containerRegistry string
		appName           string
		path              string
		serviceAccount    string
		actionVerb        string
		wantErr           error
		servingFactoryErr error
		serviceCreateErr  error
		logErr            error

		// AppLister
		deployedApps []string
		listerErr    error
	}{
		"pushes app to a configured namespace": {
			namespace:         "some-namespace",
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			appName:           "some-app",
		},
		"pushes app to default namespace": {
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			appName:           "some-app",
		},
		"app already exists, update": {
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			appName:           "some-app",
			actionVerb:        "update",
			deployedApps:      []string{"some-other-app", "some-app"},
		},
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
		"serving factory error, returns error": {
			wantErr:           errors.New("some error"),
			servingFactoryErr: errors.New("some error"),
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			appName:           "some-app",
		},
		"service create error, returns error": {
			wantErr:           errors.New("some error"),
			serviceCreateErr:  errors.New("some error"),
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			appName:           "some-app",
		},
		"service list error, returns error": {
			wantErr:           errors.New("some error"),
			listerErr:         errors.New("some error"),
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			appName:           "some-app",
		},
		"fetching logs returns an error, no error": {
			logErr:            errors.New("some error"),
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
			appName:           "some-app",
			wantErr:           errors.New("some error"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			if tc.actionVerb == "" {
				tc.actionVerb = "create"
			}

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
				testPushReaction(t, action, expectedNamespace, tc.appName, tc.containerRegistry, tc.serviceAccount, tc.actionVerb)

				// Set the ResourceVersion
				obj := action.(ktesting.CreateAction).GetObject()
				obj.(*serving.Service).ResourceVersion = tc.appName + "-version"

				return true, obj, tc.serviceCreateErr
			}))

			ctrl := gomock.NewController(t)
			fakeAppLister := kffake.NewFakeLister(ctrl)

			fakeAppListerRecorder := fakeAppLister.
				EXPECT().
				List(gomock.Any()).
				DoAndReturn(func(opts ...kf.ListOption) ([]serving.Service, error) {
					if namespace := kf.ListOptions(opts).Namespace(); namespace != expectedNamespace {
						t.Fatalf("expected namespace %s, got %s", expectedNamespace, namespace)
					}

					var apps []serving.Service
					for _, appName := range tc.deployedApps {
						s := serving.Service{}
						s.Name = appName
						s.ResourceVersion = appName + "-version"
						apps = append(apps, s)
					}

					return apps, tc.listerErr
				})

			fakeLogs := kffake.NewFakeLogTailer(ctrl)

			fakeLogTailerRecorder := fakeLogs.
				EXPECT().
				Tail(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(out io.Writer, resourceVersion, namespace string) error {
					if out == nil {
						t.Fatal("expected out to not be nil")
					}
					if namespace != expectedNamespace {
						t.Fatalf("expected namespace %s, got %s", expectedNamespace, namespace)
					}

					logResourceVersion := tc.appName + "-version"
					if resourceVersion != logResourceVersion {
						t.Fatalf("expected resourceVersion %s, got %s", logResourceVersion, resourceVersion)
					}

					logResourceVersion = resourceVersion

					return tc.logErr
				})

			var (
				imageDir string
				imageTag string
			)

			p := kf.NewPusher(
				fakeAppLister,
				func() (cserving.ServingV1alpha1Interface, error) {
					return fake, tc.servingFactoryErr
				},
				func(dir, tag string) error {
					imageDir = dir
					imageTag = tag
					return nil
				},
				fakeLogs,
			)

			var opts []kf.PushOption
			if tc.namespace != "" {
				opts = append(opts, kf.WithPushNamespace(tc.namespace))
			}
			if tc.containerRegistry != "" {
				opts = append(opts, kf.WithPushContainerRegistry(tc.containerRegistry))
			}
			if tc.serviceAccount != "" {
				opts = append(opts, kf.WithPushServiceAccount(tc.serviceAccount))
			}
			if tc.path != "" {
				opts = append(opts, kf.WithPushPath(tc.path))
			}

			gotErr := p.Push(tc.appName, opts...)
			if tc.wantErr != nil || gotErr != nil {
				// We don't really care if these were invoked if we want an
				// error.
				fakeAppListerRecorder.AnyTimes()
				fakeLogTailerRecorder.AnyTimes()

				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				return
			}

			if !reactorCalled {
				t.Fatal("Reactor was not invoked")
			}

			if imageDir != expectedPath {
				t.Fatalf("wanted cwd %s, got %s", expectedPath, imageDir)
			}

			if !strings.HasPrefix(imageTag, tc.containerRegistry) {
				t.Fatalf("want container registry prefix %s, got %s", tc.containerRegistry, imageTag)
			}

			if !strings.HasSuffix(imageTag, "latest") {
				t.Fatalf("want container registry suffix %s, got %s", "latest", imageTag)
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
