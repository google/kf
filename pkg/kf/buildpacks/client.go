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
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"time"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/doctor"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/kf"
	"github.com/buildpack/lifecycle"
	"github.com/buildpack/pack"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	gcrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	toml "github.com/pelletier/go-toml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Client is the main interface for interacting with Buildpacks.
type Client interface {
	doctor.Diagnosable

	// UploadBuildTemplate uploads a buildpack build template with the name
	// "buildpack".
	UploadBuildTemplate(imageName string) error

	// Create creates and publishes a builder image.
	Create(dir, containerRegistry string) (string, error)

	// List lists the buildpacks available.
	List() ([]string, error)
}

// BuilderFactory creates and publishes new builde image.
type BuilderFactoryCreate func(flags pack.CreateBuilderFlags) error

// RemoteImageFetcher is implemented by
// github.com/google/go-containerregistry/pkg/v1/remote.Image
type RemoteImageFetcher func(ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error)

type client struct {
	build         cbuild.BuildV1alpha1Interface
	imageFetcher  RemoteImageFetcher
	builderCreate BuilderFactoryCreate
}

// NewClient creates a new Client.
func NewClient(
	b cbuild.BuildV1alpha1Interface,
	imageFetcher RemoteImageFetcher,
	builderCreate BuilderFactoryCreate,
) Client {
	return &client{
		build:         b,
		imageFetcher:  imageFetcher,
		builderCreate: builderCreate,
	}
}

// List lists the available buildpacks.
func (c *client) List() ([]string, error) {
	templates, err := c.build.ClusterBuildTemplates().List(metav1.ListOptions{
		FieldSelector: "metadata.name=buildpack",
	})
	if err != nil {
		return nil, err
	}

	if len(templates.Items) == 0 {
		return nil, nil
	}

	builderImage := c.fetchBuilderImageName(templates.Items[0].Spec.Parameters)

	imageRef, err := name.ParseReference(builderImage, name.WeakValidation)
	if err != nil {
		return nil, err
	}

	image, err := c.imageFetcher(imageRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, err
	}

	ls, err := image.Layers()
	if err != nil {
		return nil, err
	}

	for i := len(ls) - 1; i >= 0; i-- {
		layer := ls[i]
		tr, closer, err := c.fetchImageTar(layer)
		if err != nil {
			return nil, err
		}
		defer closer.Close()

		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			if header.Name == "/buildpacks/order.toml" {
				return c.readOrder(tr)
			}
		}
	}

	return nil, nil
}

func (c *client) readOrder(reader io.Reader) ([]string, error) {
	var buildpackIDs []string
	var order struct {
		Groups []lifecycle.BuildpackGroup `toml:"groups"`
	}
	if err := toml.NewDecoder(reader).Decode(&order); err != nil {
		return nil, err
	}

	for _, group := range order.Groups {
		for _, bp := range group.Buildpacks {
			buildpackIDs = append(buildpackIDs, bp.ID)
		}
	}

	return buildpackIDs, nil
}

func (c *client) fetchImageTar(layer gcrv1.Layer) (*tar.Reader, io.Closer, error) {
	ucl, err := layer.Uncompressed()
	if err != nil {
		return nil, nil, err
	}

	return tar.NewReader(ucl), ucl, nil
}

func (c *client) fetchBuilderImageName(params []build.ParameterSpec) string {
	for _, p := range params {
		if p.Name == "BUILDER_IMAGE" && p.Default != nil {
			return *p.Default
		}
	}

	return ""
}

// UploadBuildTemplate uploads a buildpack build template with the name
// "buildpack".
func (c *client) UploadBuildTemplate(imageName string) error {
	if imageName == "" {
		return errors.New("image name must not be empty")
	}

	// TODO: It would be nice if we generated this instead.
	if _, err := c.deployer()(&build.ClusterBuildTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "build.knative.dev/v1alpha1",
			Kind:       "ClusterBuildTemplate",
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
					Default:     c.strToPtr("packs/run:v3alpha2"),
				},
				{
					Name:        "BUILDER_IMAGE",
					Description: `The builder image (must include v3 lifecycle and compatible buildpacks).`,
					Default:     c.strToPtr(imageName),
				},
				{
					Name:        "USE_CRED_HELPERS",
					Description: `Use Docker credential helpers for Google's GCR, Amazon's ECR, or Microsoft's ACR.`,
					Default:     c.strToPtr("true"),
				},
				{
					Name:        "CACHE",
					Description: `The name of the persistent app cache volume`,
					Default:     c.strToPtr("empty-dir"),
				},
				{
					Name:        "USER_ID",
					Description: `The user ID of the builder image user`,
					Default:     c.strToPtr("1000"),
				},
				{
					Name:        "GROUP_ID",
					Description: `The group ID of the builder image user`,
					Default:     c.strToPtr("1000"),
				},
				{
					Name:        "BUILDPACK",
					Description: `When set, skip the detect step and use the given buildpack.`,
					Default:     c.strToPtr(""),
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

type deployer func(*build.ClusterBuildTemplate) (*build.ClusterBuildTemplate, error)

func (c *client) deployer() deployer {
	builds, err := c.build.ClusterBuildTemplates().List(metav1.ListOptions{
		FieldSelector: "metadata.name=buildpack",
	})

	if err != nil {
		// Simplify workflow and just return a deployer that will fail with the
		// given error.
		return func(t *build.ClusterBuildTemplate) (*build.ClusterBuildTemplate, error) {
			return nil, err
		}
	}

	if len(builds.Items) == 0 {
		return func(t *build.ClusterBuildTemplate) (*build.ClusterBuildTemplate, error) {
			return c.build.ClusterBuildTemplates().Create(t)
		}
	}

	return func(t *build.ClusterBuildTemplate) (*build.ClusterBuildTemplate, error) {
		t.ResourceVersion = builds.Items[0].ResourceVersion
		return c.build.ClusterBuildTemplates().Update(t)
	}
}

func (c *client) strToPtr(s string) *string {
	return &s
}

// Create creates and publishes a builder image.
func (c *client) Create(dir, containerRegistry string) (string, error) {
	if dir == "" {
		return "", kf.ConfigErr{Reason: "dir must not be empty"}
	}
	if containerRegistry == "" {
		return "", kf.ConfigErr{Reason: "containerRegistry must not be empty"}
	}

	if filepath.Base(dir) != "builder.toml" {
		dir = filepath.Join(dir, "builder.toml")
	}

	imageName := path.Join(containerRegistry, fmt.Sprintf("buildpack-builder:%d", time.Now().UnixNano()))
	if err := c.builderCreate(pack.CreateBuilderFlags{
		Publish:         true,
		BuilderTomlPath: dir,
		RepoName:        imageName,
	}); err != nil {
		return "", err
	}

	return imageName, nil
}

// Diagnose checks to see if the cluster has buildpacks.
func (c *client) Diagnose(d *doctor.Diagnostic) {
	d.Run("Buildpacks", func(d *doctor.Diagnostic) {
		buildpacks, err := c.List()
		if err != nil {
			d.Fatalf("Error fetching Buildpacks: %s", err)
		}
		if len(buildpacks) == 0 {
			d.Fatal("Expected to find at least one buildpack")
		}
	})
}
