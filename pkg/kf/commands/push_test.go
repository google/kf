package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestPushCommand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name              string
		args              []string
		namespace         string
		containerRegistry string
		serviceAccount    string
		wantErr           error
		wantUsageOnErr    bool
		servingFactoryErr error
		serviceCreateErr  error
	}{
		{
			name:              "uses configured namespace",
			namespace:         "some-namespace",
			args:              []string{"app-name"},
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
		},
		{
			name:           "container registry not configured, returns error",
			args:           []string{"app-name"},
			wantErr:        errors.New("container registry is not set"),
			serviceAccount: "some-service-account",
			wantUsageOnErr: true,
		},
		{
			name:              "service account not configured, returns error",
			args:              []string{"app-name"},
			wantErr:           errors.New("service account is not set"),
			containerRegistry: "some-reg.io",
			wantUsageOnErr:    true,
		},
		{
			name:              "serving factory error",
			args:              []string{"app-name"},
			wantErr:           errors.New("some error"),
			servingFactoryErr: errors.New("some error"),
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
		},
		{
			name:              "service create error",
			args:              []string{"app-name"},
			wantErr:           errors.New("some error"),
			serviceCreateErr:  errors.New("some error"),
			containerRegistry: "some-reg.io",
			serviceAccount:    "some-service-account",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			fake := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}

			called := false
			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				called = true
				testPushReaction(t, action, tc.namespace, tc.args[0], tc.containerRegistry, tc.serviceAccount)
				return tc.serviceCreateErr != nil, nil, tc.serviceCreateErr
			}))

			var (
				imageDir string
				imageTag string
			)

			c := NewPushCommand(&KfParams{
				Namespace: tc.namespace,
				Output:    buffer,
				ServingFactory: func() (cserving.ServingV1alpha1Interface, error) {
					return fake, tc.servingFactoryErr
				},
			},
				func(dir, tag string) error {
					imageDir = dir
					imageTag = tag
					return nil
				},
			)

			c.Flags().Set("container-registry", tc.containerRegistry)
			c.Flags().Set("service-account", tc.serviceAccount)
			gotErr := c.RunE(c, tc.args)
			if tc.wantErr != nil || gotErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				if !tc.wantUsageOnErr != c.SilenceUsage {
					t.Fatalf("wanted %v, got %v", !tc.wantUsageOnErr, c.SilenceUsage)
				}

				return
			}

			if !called {
				t.Fatal("Reactor was not invoked")
			}

			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			if imageDir != cwd {
				t.Fatalf("wanted cwd %s, got %s", cwd, imageDir)
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
) {
	t.Helper()

	if action.GetNamespace() != namespace {
		t.Fatalf("wanted namespace: %s, got: %s", namespace, action.GetNamespace())
	}

	if !action.Matches("create", "services") {
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
