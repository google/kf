package buildpacks_test

import (
	"errors"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	"github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestBuildTemplateUploader(t *testing.T) {
	t.Parallel()

	for tn, tc := range map[string]struct {
		ImageName            string
		ExpectedErr          error
		BuildFactoryErr      error
		BuildTemplateErr     error
		BuildTemplateItems   []string
		ListBuildTemplateErr error
		Opts                 []buildpacks.UploadBuildTemplateOption
		HandleDeployAction   func(t *testing.T, action ktesting.Action)
		HandleListAction     func(t *testing.T, action ktesting.Action)
	}{
		"sets meta data for list": {
			ImageName: "some-image",
			HandleListAction: func(t *testing.T, action ktesting.Action) {
				testutil.AssertEqual(t, "Verb", "list", action.GetVerb())
				testutil.AssertEqual(t, "Resource", "buildtemplates", action.GetResource().Resource)
				testutil.AssertEqual(t, "namespace", "default", action.GetNamespace())
				testutil.AssertEqual(t, "FieldSelector Field", "metadata.name", action.(ktesting.ListActionImpl).ListRestrictions.Fields.Requirements()[0].Field)
				testutil.AssertEqual(t, "FieldSelector Value", "buildpack", action.(ktesting.ListActionImpl).ListRestrictions.Fields.Requirements()[0].Value)
			},
		},
		"sets meta data for create": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				testutil.AssertEqual(t, "Verb", "create", action.GetVerb())
				testutil.AssertEqual(t, "Resource", "buildtemplates", action.GetResource().Resource)

				bt := action.(ktesting.CreateActionImpl).Object.(*build.BuildTemplate)
				testutil.AssertEqual(t, "apiVersion", "build.knative.dev/v1alpha1", bt.APIVersion)
				testutil.AssertEqual(t, "kind", "BuildTemplate", bt.Kind)
				testutil.AssertEqual(t, "Name", "buildpack", bt.Name)
			},
		},
		"sets meta data for update": {
			ImageName:          "some-image",
			BuildTemplateItems: []string{"template-1"},
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				testutil.AssertEqual(t, "Verb", "update", action.GetVerb())
				testutil.AssertEqual(t, "Resource", "buildtemplates", action.GetResource().Resource)

				bt := action.(ktesting.UpdateActionImpl).Object.(*build.BuildTemplate)
				testutil.AssertEqual(t, "Resource", "template-1-version", bt.ResourceVersion)
			},
		},
		"sets the step parameters": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				bt := action.(ktesting.CreateActionImpl).Object.(*build.BuildTemplate)
				params := map[interface{}]interface{}{}
				for _, p := range bt.Spec.Parameters {
					if p.Default == nil {
						params[p.Name] = ""
						continue
					}

					params[p.Name] = *p.Default
				}

				testutil.AssertKeyWithValue(t, params, "IMAGE", "")
				testutil.AssertKeyWithValue(t, params, "RUN_IMAGE", "packs/run:v3alpha2")
				testutil.AssertKeyWithValue(t, params, "BUILDER_IMAGE", "some-image")
				testutil.AssertKeyWithValue(t, params, "USE_CRED_HELPERS", "true")
				testutil.AssertKeyWithValue(t, params, "CACHE", "empty-dir")
				testutil.AssertKeyWithValue(t, params, "USER_ID", "1000")
				testutil.AssertKeyWithValue(t, params, "GROUP_ID", "1000")
			},
		},
		"step prepare": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				step := action.(ktesting.CreateActionImpl).Object.(*build.BuildTemplate).Spec.Steps[0]
				testutil.AssertEqual(t, "Name", "prepare", step.Name)
				testutil.AssertEqual(t, "Image", "alpine", step.Image)
				testutil.AssertEqual(t, "Command", []string{"/bin/sh"}, step.Command)
				testutil.AssertEqual(t, "Args", []string{
					"-c",
					`chown -R "${USER_ID}:${GROUP_ID}" "/builder/home" &&
						 chown -R "${USER_ID}:${GROUP_ID}" /layers &&
						 chown -R "${USER_ID}:${GROUP_ID}" /workspace`,
				}, step.Args)
				testutil.AssertEqual(t, "VolumeMounts.Name", "${CACHE}", step.VolumeMounts[0].Name)
				testutil.AssertEqual(t, "VolumeMounts.MountPath", "/layers", step.VolumeMounts[0].MountPath)
				testutil.AssertEqual(t, "ImagePullPolicy", "Always", string(step.ImagePullPolicy))
			},
		},
		"step detect": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				step := action.(ktesting.CreateActionImpl).Object.(*build.BuildTemplate).Spec.Steps[1]
				testutil.AssertEqual(t, "Name", "detect", step.Name)
				testutil.AssertEqual(t, "Image", "${BUILDER_IMAGE}", step.Image)
				testutil.AssertEqual(t, "Command", []string{"/lifecycle/detector"}, step.Command)
				testutil.AssertEqual(t, "Args", []string{
					"-app=/workspace",
					"-group=/layers/group.toml",
					"-plan=/layers/plan.toml",
				}, step.Args)
				testutil.AssertEqual(t, "VolumeMounts.Name", "${CACHE}", step.VolumeMounts[0].Name)
				testutil.AssertEqual(t, "VolumeMounts.MountPath", "/layers", step.VolumeMounts[0].MountPath)
				testutil.AssertEqual(t, "ImagePullPolicy", "Always", string(step.ImagePullPolicy))
			},
		},
		"step analyze": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				step := action.(ktesting.CreateActionImpl).Object.(*build.BuildTemplate).Spec.Steps[2]
				testutil.AssertEqual(t, "Name", "analyze", step.Name)
				testutil.AssertEqual(t, "Image", "${BUILDER_IMAGE}", step.Image)
				testutil.AssertEqual(t, "Command", []string{"/lifecycle/analyzer"}, step.Command)
				testutil.AssertEqual(t, "Args", []string{
					"-layers=/layers",
					"-helpers=${USE_CRED_HELPERS}",
					"-group=/layers/group.toml",
					"${IMAGE}",
				}, step.Args)
				testutil.AssertEqual(t, "VolumeMounts.Name", "${CACHE}", step.VolumeMounts[0].Name)
				testutil.AssertEqual(t, "VolumeMounts.MountPath", "/layers", step.VolumeMounts[0].MountPath)
				testutil.AssertEqual(t, "ImagePullPolicy", "Always", string(step.ImagePullPolicy))
			},
		},
		"step build": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				step := action.(ktesting.CreateActionImpl).Object.(*build.BuildTemplate).Spec.Steps[3]
				testutil.AssertEqual(t, "Name", "build", step.Name)
				testutil.AssertEqual(t, "Image", "${BUILDER_IMAGE}", step.Image)
				testutil.AssertEqual(t, "Command", []string{"/lifecycle/builder"}, step.Command)
				testutil.AssertEqual(t, "Args", []string{
					"-layers=/layers",
					"-app=/workspace",
					"-group=/layers/group.toml",
					"-plan=/layers/plan.toml",
				}, step.Args)
				testutil.AssertEqual(t, "VolumeMounts.Name", "${CACHE}", step.VolumeMounts[0].Name)
				testutil.AssertEqual(t, "VolumeMounts.MountPath", "/layers", step.VolumeMounts[0].MountPath)
				testutil.AssertEqual(t, "ImagePullPolicy", "Always", string(step.ImagePullPolicy))
			},
		},
		"step export": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				step := action.(ktesting.CreateActionImpl).Object.(*build.BuildTemplate).Spec.Steps[4]
				testutil.AssertEqual(t, "Name", "export", step.Name)
				testutil.AssertEqual(t, "Image", "${BUILDER_IMAGE}", step.Image)
				testutil.AssertEqual(t, "Command", []string{"/lifecycle/exporter"}, step.Command)
				testutil.AssertEqual(t, "Args", []string{
					"-layers=/layers",
					"-helpers=${USE_CRED_HELPERS}",
					"-app=/workspace",
					"-image=${RUN_IMAGE}",
					"-group=/layers/group.toml",
					"${IMAGE}",
				}, step.Args)
				testutil.AssertEqual(t, "VolumeMounts.Name", "${CACHE}", step.VolumeMounts[0].Name)
				testutil.AssertEqual(t, "VolumeMounts.MountPath", "/layers", step.VolumeMounts[0].MountPath)
				testutil.AssertEqual(t, "ImagePullPolicy", "Always", string(step.ImagePullPolicy))
			},
		},
		"volumes": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				volumes := action.(ktesting.CreateActionImpl).Object.(*build.BuildTemplate).Spec.Volumes
				testutil.AssertEqual(t, "Volumes.Name", "empty-dir", volumes[0].Name)
			},
		},
		"default namespace": {
			ImageName: "some-image",
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				testutil.AssertEqual(t, "Namespace", "default", action.GetNamespace())
			},
		},
		"uses custom namespace": {
			ImageName: "some-image",
			Opts: []buildpacks.UploadBuildTemplateOption{
				buildpacks.WithUploadBuildTemplateNamespace("some-namespace"),
			},
			HandleDeployAction: func(t *testing.T, action ktesting.Action) {
				testutil.AssertEqual(t, "Namespace", "some-namespace", action.GetNamespace())
			},
		},
		"empty image name returns an error": {
			ImageName:   "",
			ExpectedErr: errors.New("image name must not be empty"),
		},
		"upload build template error": {
			ImageName:        "some-image",
			BuildTemplateErr: errors.New("some-error"),
			ExpectedErr:      errors.New("some-error"),
		},
		"list build templates error": {
			ImageName:            "some-image",
			ListBuildTemplateErr: errors.New("some-error"),
			ExpectedErr:          errors.New("some-error"),
		},
		"build factory error": {
			ImageName:       "some-image",
			BuildFactoryErr: errors.New("some-error"),
			ExpectedErr:     errors.New("some-error"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			fake := &fake.FakeBuildV1alpha1{
				Fake: &ktesting.Fake{},
			}
			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if action.GetVerb() == "list" {
					if tc.HandleListAction != nil {
						tc.HandleListAction(t, action)
					}
					if tc.ListBuildTemplateErr != nil {
						return true, nil, tc.ListBuildTemplateErr
					}

					if len(tc.BuildTemplateItems) > 0 {
						var bt build.BuildTemplateList
						for _, name := range tc.BuildTemplateItems {
							bt.Items = append(bt.Items, build.BuildTemplate{
								ObjectMeta: metav1.ObjectMeta{
									Name:            name,
									ResourceVersion: name + "-version",
								},
							})
						}

						return true, &bt, nil
					}
				}

				if action.GetVerb() == "create" || action.GetVerb() == "update" {
					if tc.HandleDeployAction != nil {
						tc.HandleDeployAction(t, action)
					}
					return tc.BuildTemplateErr != nil, nil, tc.BuildTemplateErr
				}

				return false, nil, nil
			}))

			u := buildpacks.NewBuildTemplateUploader(func() (cbuild.BuildV1alpha1Interface, error) {
				return fake, tc.BuildFactoryErr
			})

			gotErr := u.UploadBuildTemplate(tc.ImageName, tc.Opts...)
			if gotErr != nil || tc.ExpectedErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, gotErr)
				return
			}
		})
	}
}
