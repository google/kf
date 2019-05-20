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

package buildpacks

import (
	"errors"

	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go run ../internal/tools/option-builder/option-builder.go options.yml options.go

// BuildTemplateUploader uploads a build template
type BuildTemplateUploader interface {
	// UploadBuildTemplate uploads a buildpack build template with the name
	// "buildpack".
	UploadBuildTemplate(imageName string, opts ...UploadBuildTemplateOption) error
}

// buildTemplateUploader uploads a new buildpack build template. It should be
// created via NewBuildTemplateUploader.
type buildTemplateUploader struct {
	c cbuild.BuildV1alpha1Interface
}

// NewBuildTemplateUploader creates a new BuildTemplateUploader.
func NewBuildTemplateUploader(c cbuild.BuildV1alpha1Interface) BuildTemplateUploader {
	return &buildTemplateUploader{
		c: c,
	}
}

// UploadBuildTemplate uploads a buildpack build template with the name
// "buildpack".
func (u *buildTemplateUploader) UploadBuildTemplate(imageName string, opts ...UploadBuildTemplateOption) error {
	if imageName == "" {
		return errors.New("image name must not be empty")
	}

	cfg := UploadBuildTemplateOptionDefaults().Extend(opts).toConfig()

	// TODO: It would be nice if we generated this instead.
	if _, err := u.deployer(cfg.Namespace)(&build.BuildTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "build.knative.dev/v1alpha1",
			Kind:       "BuildTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "buildpack",
		},
		Spec: build.BuildTemplateSpec{
			Parameters: []build.ParameterSpec{
				{
					Name:        "IMAGE",
					Description: `The image you wish to create. For example, "repo/example", or "example.com/repo/image"`,
				},
				{
					Name:        "RUN_IMAGE",
					Description: `The run image buildpacks will use as the base for IMAGE.`,
					Default:     u.strToPtr("packs/run:v3alpha2"),
				},
				{
					Name:        "BUILDER_IMAGE",
					Description: `The builder image (must include v3 lifecycle and compatible buildpacks).`,
					Default:     u.strToPtr(imageName),
				},
				{
					Name:        "USE_CRED_HELPERS",
					Description: `Use Docker credential helpers for Google's GCR, Amazon's ECR, or Microsoft's ACR.`,
					Default:     u.strToPtr("true"),
				},
				{
					Name:        "CACHE",
					Description: `The name of the persistent app cache volume`,
					Default:     u.strToPtr("empty-dir"),
				},
				{
					Name:        "USER_ID",
					Description: `The user ID of the builder image user`,
					Default:     u.strToPtr("1000"),
				},
				{
					Name:        "GROUP_ID",
					Description: `The group ID of the builder image user`,
					Default:     u.strToPtr("1000"),
				},
				{
					Name:        "BUILDPACK",
					Description: `When set, skip the detect step and use the given buildpack.`,
					Default:     u.strToPtr(""),
				},
			},
			Steps: []corev1.Container{
				{
					Name:    "prepare",
					Image:   "alpine",
					Command: []string{"/bin/sh"},
					Args: []string{
						"-c",
						`chown -R "${USER_ID}:${GROUP_ID}" "/builder/home" &&
						 chown -R "${USER_ID}:${GROUP_ID}" /layers &&
						 chown -R "${USER_ID}:${GROUP_ID}" /workspace`,
					},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "${CACHE}",
						MountPath: "/layers",
					}},
					ImagePullPolicy: "Always",
				},
				{
					Name:    "detect",
					Image:   "${BUILDER_IMAGE}",
					Command: []string{"/bin/bash"},
					Args: []string{
						"-c",
						`if [[ -z "${BUILDPACK}" ]]; then
  /lifecycle/detector \
  -app=/workspace \
  -group=/layers/group.toml \
  -plan=/layers/plan.toml
else
touch /layers/plan.toml
cat <<EOF > /layers/group.toml
[[buildpacks]]
  id = "${BUILDPACK}"
  version = "latest"
EOF
fi`,
					},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "${CACHE}",
						MountPath: "/layers",
					}},
					ImagePullPolicy: "Always",
				},
				{
					Name:    "analyze",
					Image:   "${BUILDER_IMAGE}",
					Command: []string{"/lifecycle/analyzer"},
					Args: []string{
						"-layers=/layers",
						"-helpers=${USE_CRED_HELPERS}",
						"-group=/layers/group.toml",
						"${IMAGE}",
					},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "${CACHE}",
						MountPath: "/layers",
					}},
					ImagePullPolicy: "Always",
				},
				{
					Name:    "build",
					Image:   "${BUILDER_IMAGE}",
					Command: []string{"/lifecycle/builder"},
					Args: []string{
						"-layers=/layers",
						"-app=/workspace",
						"-group=/layers/group.toml",
						"-plan=/layers/plan.toml",
					},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "${CACHE}",
						MountPath: "/layers",
					}},
					ImagePullPolicy: "Always",
				},
				{
					Name:    "export",
					Image:   "${BUILDER_IMAGE}",
					Command: []string{"/lifecycle/exporter"},
					Args: []string{
						"-layers=/layers",
						"-helpers=${USE_CRED_HELPERS}",
						"-app=/workspace",
						"-image=${RUN_IMAGE}",
						"-group=/layers/group.toml",
						"${IMAGE}",
					},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "${CACHE}",
						MountPath: "/layers",
					}},
					ImagePullPolicy: "Always",
				},
			},
			Volumes: []corev1.Volume{{
				Name: "empty-dir",
			}},
		},
	}); err != nil {
		return err
	}

	return nil
}

type deployer func(*build.BuildTemplate) (*build.BuildTemplate, error)

func (u *buildTemplateUploader) deployer(namespace string) deployer {
	builds, err := u.c.BuildTemplates(namespace).List(metav1.ListOptions{
		FieldSelector: "metadata.name=buildpack",
	})

	if err != nil {
		// Simplify workflow and just return a deployer that will fail with the
		// given error.
		return func(t *build.BuildTemplate) (*build.BuildTemplate, error) {
			return nil, err
		}
	}

	if len(builds.Items) == 0 {
		return func(t *build.BuildTemplate) (*build.BuildTemplate, error) {
			return u.c.BuildTemplates(namespace).Create(t)
		}
	}

	return func(t *build.BuildTemplate) (*build.BuildTemplate, error) {
		t.ResourceVersion = builds.Items[0].ResourceVersion
		return u.c.BuildTemplates(namespace).Update(t)
	}
}

func (u *buildTemplateUploader) strToPtr(s string) *string {
	return &s
}
