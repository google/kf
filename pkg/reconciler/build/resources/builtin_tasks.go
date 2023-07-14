// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/tektonutil"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// FindBuiltinTask returns a TaskSpec for a build task that's built-in to Kf.
// The implementation details of these tasks may change because they're not
// public.
// If no task matching the given ref is found, then nil is returned.
func FindBuiltinTask(cfg *config.DefaultsConfig, buildSpec v1alpha1.BuildSpec, googleServiceAccount string) *tektonv1beta1.TaskSpec {
	if cfg == nil {
		return nil
	}

	buildTaskRef := buildSpec.BuildTaskRef

	if buildTaskRef.Kind != v1alpha1.BuiltinTaskKind {
		return nil
	}

	if buildTaskRef.APIVersion != v1alpha1.BuiltinTaskAPIVersion {
		return nil
	}

	switch buildTaskRef.Name {
	case v1alpha1.BuildpackV2BuildTaskName:
		return buildpackV2Task(cfg)
	case v1alpha1.DockerfileBuildTaskName:
		return dockerfileBuildTask(cfg)
	case v1alpha1.BuildpackV3BuildTaskName:
		return buildpackV3Build(cfg, buildSpec, googleServiceAccount)
	}

	return nil
}

func buildTaskResults() []tektonv1beta1.TaskResult {
	return []tektonv1beta1.TaskResult{
		{
			Name:        v1alpha1.TaskRunParamDestinationImage,
			Description: "image built by buildpacks",
			Type:        tektonv1beta1.ResultsTypeString,
		},
	}
}

func buildpackV2Task(cfg *config.DefaultsConfig) *tektonv1beta1.TaskSpec {
	var resources corev1.ResourceRequirements
	if cfg.BuildPodResources != nil {
		resources = *cfg.BuildPodResources
	}

	return &tektonv1beta1.TaskSpec{
		Params: []tektonv1beta1.ParamSpec{
			tektonutil.DefaultStringParam("BUILD_NAME", "The name of the Build to push destination image for.", ""),
			tektonutil.DefaultStringParam(v1alpha1.TaskRunParamDestinationImage, "The URI that'll be used for the application's output image.", ""),
			tektonutil.DefaultStringParam("SOURCE_IMAGE", "The image that contains the app's source code.", ""),
			tektonutil.DefaultStringParam("SOURCE_PACKAGE_NAMESPACE", "The namespace of the source package.", ""),
			tektonutil.DefaultStringParam("SOURCE_PACKAGE_NAME", "The name of the source package.", ""),
			tektonutil.StringParam("BUILDPACKS", "Ordered list of comma separated builtpacks to attempt."),
			tektonutil.StringParam("RUN_IMAGE", "The run image apps will use as the base for IMAGE (output)."),
			tektonutil.StringParam("BUILDER_IMAGE", "The image on which builds will run."),
			tektonutil.DefaultStringParam("SKIP_DETECT", "Skip the detect phase", "false"),
		},
		Results: buildTaskResults(),
		Steps: []tektonv1beta1.Step{
			{
				Name:    "source-extraction",
				Image:   cfg.BuildHelpersImage,
				Command: []string{"/ko-app/build-helpers"},
				Args: []string{
					"extract",
					"--output-dir",
					"/staging/app",
					"--source-package-namespace",
					"$(inputs.params.SOURCE_PACKAGE_NAMESPACE)",
					"--source-package-name",
					"$(inputs.params.SOURCE_PACKAGE_NAME)",
					"--source-image",
					"$(inputs.params.SOURCE_IMAGE)",
				},
				VolumeMounts: []corev1.VolumeMount{
					{Name: "staging-tmp-dir", MountPath: "/staging"},
				},
			},
			{
				Name:    "copy-lifecycle",
				Image:   cfg.BuildpacksV2LifecycleImage,
				Command: []string{"/ko-app/installer"},
				VolumeMounts: []corev1.VolumeMount{
					{Name: "staging-tmp-dir", MountPath: "/staging"},
				},
			},
			{
				Name:  "run-lifecycle",
				Image: "$(inputs.params.BUILDER_IMAGE)",
				// NOTE: this command shouldn't be run as root, instead it should be run as
				// vcap:vcap
				Command: []string{"bash"},
				// A /tmp directory is necessary because some buildpacks use /tmp
				// which causes cross-device links to be made because Tekton mounts
				// the /workspace directory.
				// TODO: add fuseimage to buildpackv2Build buildspec, and pull from config.
				Args: []string{
					"-euc",
					`
cp -r /staging/app /tmp/app
/workspace/builder \
  -buildArtifactsCacheDir=/tmp/cache \
  -buildDir=/tmp/app \
  -buildpacksDir=/tmp/buildpacks \
  -outputBuildArtifactsCache=/tmp/output-cache \
  -outputDroplet=/tmp/droplet \
  -outputMetadata=/tmp/result.json \
  "-buildpackOrder=$(inputs.params.BUILDPACKS)" \
  "-skipDetect=$(inputs.params.SKIP_DETECT)"
cp -r /tmp/droplet /workspace/droplet

cat << 'EOF' > /workspace/entrypoint.bash
#!/usr/bin/env bash
set -e

if [[ "$@" == "" ]]; then
  exec /lifecycle/launcher "/home/vcap/app" "" ""
else
  exec /lifecycle/launcher "/home/vcap/app" "$@" ""
fi

EOF
chmod a+x /workspace/entrypoint.bash

start_cmd=$(cat /tmp/result.json | jq .process_types.web)

cat << EOF > /workspace/Dockerfile
FROM gcr.io/kf-releases/fusesidecar:v2.11.14 as builder

FROM $(inputs.params.RUN_IMAGE)
COPY launcher /lifecycle/launcher
COPY entrypoint.bash /lifecycle/entrypoint.bash
COPY --from=builder --chown=root:vcap /bin/mapfs /bin/mapfs
LABEL StartCommand=$start_cmd

# need this to allow users other than root to use fuse.
RUN echo "user_allow_other" >> /etc/fuse.conf
RUN chmod 644 /etc/fuse.conf

RUN chmod 750 /bin/mapfs
# so that whoever runs this has the privileges of the owner(root).
RUN chmod u+s /bin/mapfs

WORKDIR /home/vcap
USER vcap:vcap
COPY droplet droplet.tar.gz
RUN tar -xzf droplet.tar.gz && rm droplet.tar.gz
ENTRYPOINT ["/lifecycle/entrypoint.bash"]
EOF
`,
				},
				Resources: resources,
				VolumeMounts: []corev1.VolumeMount{
					{Name: "staging-tmp-dir", MountPath: "/staging"},
				},
			},

			{
				Name:       "build",
				WorkingDir: "/workspace",
				Command:    []string{"/kaniko/executor"},
				Image:      cfg.BuildKanikoExecutorImage,
				Args: []string{
					"--dockerfile",
					"/workspace/Dockerfile",
					"--context",
					"/workspace",
					"--destination",
					"$(inputs.params.DESTINATION_IMAGE)",
					"--oci-layout-path",
					"/tekton/home/image-outputs/IMAGE",
					"--single-snapshot",
					"--no-push",
					"--tarPath",
					"/workspace/image.tar",
				},
				Resources: resources,
				VolumeMounts: []corev1.VolumeMount{
					{Name: "cache-dir", MountPath: "/cache"},
					{Name: "staging-tmp-dir", MountPath: "/workspace/staging"},
				},
			},
			{
				Name:       "publish",
				WorkingDir: "/workspace",
				Command:    []string{"/ko-app/build-helpers"},
				Image:      cfg.BuildHelpersImage,
				Args: []string{
					"publish",
					"/workspace/image.tar",
					"$(inputs.params.SOURCE_PACKAGE_NAMESPACE)",
					"$(inputs.params.BUILD_NAME)",
				},
			},
			{
				Name:       "write results",
				WorkingDir: "/workspace",
				Command:    []string{"/ko-app/build-helpers"},
				Image:      cfg.BuildHelpersImage,
				Args: []string{
					"write-result",
					"$(inputs.params.DESTINATION_IMAGE)",
					"$(results.IMAGE.path)",
				},
			},
		},
		Volumes: []corev1.Volume{
			tektonutil.EmptyVolume("cache-dir"),
			tektonutil.EmptyVolume("staging-tmp-dir"),
		},
	}
}

func dockerfileBuildTask(cfg *config.DefaultsConfig) *tektonv1beta1.TaskSpec {
	var resources corev1.ResourceRequirements
	if cfg.BuildPodResources != nil {
		resources = *cfg.BuildPodResources
	}

	layers := []corev1.VolumeMount{
		{Name: "layers-dir", MountPath: "/layers"},
	}

	return &tektonv1beta1.TaskSpec{
		Params: []tektonv1beta1.ParamSpec{
			tektonutil.DefaultStringParam("BUILD_NAME", "The name of the Build to push destination image for.", ""),
			tektonutil.DefaultStringParam(v1alpha1.TaskRunParamDestinationImage, "The URI that'll be used for the application's output image.", ""),
			tektonutil.DefaultStringParam("SOURCE_IMAGE", "The image that contains the app's source code.", ""),
			tektonutil.DefaultStringParam("SOURCE_PACKAGE_NAMESPACE", "The namespace of the source package.", ""),
			tektonutil.DefaultStringParam("SOURCE_PACKAGE_NAME", "The name of the source package.", ""),
			tektonutil.DefaultStringParam("DOCKERFILE", "Path to the Dockerfile to build.", "./Dockerfile"),
		},
		Results: buildTaskResults(),
		Steps: []tektonv1beta1.Step{
			{
				Name:    "source-extraction",
				Image:   cfg.BuildHelpersImage,
				Command: []string{"/ko-app/build-helpers"},
				Args: []string{
					"extract",
					"--output-dir",
					"/layers/source",
					"--source-package-namespace",
					"$(inputs.params.SOURCE_PACKAGE_NAMESPACE)",
					"--source-package-name",
					"$(inputs.params.SOURCE_PACKAGE_NAME)",
					"--source-image",
					"$(inputs.params.SOURCE_IMAGE)",
				},
				VolumeMounts: layers,
			},
			{
				Name:       "build",
				WorkingDir: "/layers/source",
				Image:      cfg.BuildKanikoExecutorImage,
				Command:    []string{"/kaniko/executor"},
				Args: []string{
					"--dockerfile",
					"$(inputs.params.DOCKERFILE)",
					"--context",
					"/layers/source/",
					"--destination",
					"$(inputs.params.DESTINATION_IMAGE)",
					"--no-push",
					"--tarPath",
					"/workspace/image.tar",
				},
				Resources:    resources,
				VolumeMounts: layers,
			},
			{
				Name:       "publish",
				WorkingDir: "/workspace",
				Command:    []string{"/ko-app/build-helpers"},
				Image:      cfg.BuildHelpersImage,
				Args: []string{
					"publish",
					"/workspace/image.tar",
					"$(inputs.params.SOURCE_PACKAGE_NAMESPACE)",
					"$(inputs.params.BUILD_NAME)",
				},
				VolumeMounts: layers,
			},
			{
				Name:       "write results",
				WorkingDir: "/workspace",
				Command:    []string{"/ko-app/build-helpers"},
				Image:      cfg.BuildHelpersImage,
				Args: []string{
					"write-result",
					"$(inputs.params.DESTINATION_IMAGE)",
					"$(results.IMAGE.path)",
				},
			},
		},
		Volumes: []corev1.Volume{
			tektonutil.EmptyVolume("layers-dir"),
		},
	}
}

func buildpackV3Build(cfg *config.DefaultsConfig, buildSpec v1alpha1.BuildSpec, googleServiceAccount string) *tektonv1beta1.TaskSpec {
	var resources corev1.ResourceRequirements
	if cfg.BuildPodResources != nil {
		resources = *cfg.BuildPodResources
	}

	cacheAndLayers := []corev1.VolumeMount{
		{Name: "cache-dir", MountPath: "/cache"},
		{Name: "layers-dir", MountPath: "/layers"},
		{Name: "platform-dir", MountPath: "/platform"},
	}

	platformEnvSet := sets.NewString()

	for _, v := range buildSpec.Env {
		platformEnvSet.Insert(v.Name)
	}

	return &tektonv1beta1.TaskSpec{
		Params: []tektonv1beta1.ParamSpec{
			tektonutil.DefaultStringParam("SOURCE_IMAGE", "The image that contains the app's source code.", ""),
			tektonutil.DefaultStringParam(v1alpha1.TaskRunParamDestinationImage, "The URI that'll be used for the application's output image.", ""),
			tektonutil.DefaultStringParam("SOURCE_PACKAGE_NAMESPACE", "The namespace of the source package.", ""),
			tektonutil.DefaultStringParam("SOURCE_PACKAGE_NAME", "The name of the source package.", ""),
			tektonutil.DefaultStringParam("BUILDPACK", "When set, skip the detect step and use the given buildpack.", ""),
			tektonutil.StringParam("RUN_IMAGE", "The run image buildpacks will use as the base for IMAGE (output)."),
			tektonutil.StringParam("BUILDER_IMAGE", "The image on which builds will run (must include v3 lifecycle and compatible buildpacks)."),
		},
		Results: buildTaskResults(),
		Steps: []tektonv1beta1.Step{
			{
				Name:    "source-extraction",
				Image:   cfg.BuildHelpersImage,
				Command: []string{"/ko-app/build-helpers"},
				Args: []string{
					"extract",
					"--output-dir",
					"/layers/source",
					"--source-package-namespace",
					"$(inputs.params.SOURCE_PACKAGE_NAMESPACE)",
					"--source-package-name",
					"$(inputs.params.SOURCE_PACKAGE_NAME)",
					"--source-image",
					"$(inputs.params.SOURCE_IMAGE)",
				},
				VolumeMounts: cacheAndLayers,
			},
			{
				Name:    "info",
				Image:   cfg.BuildInfoImage,
				Command: []string{"/ko-app/setup-buildpack-build"},
				Args: []string{
					"--app",
					"/layers/source",
					"--image",
					"$(outputs.resources.IMAGE.url)",
					"--run-image",
					"$(inputs.params.RUN_IMAGE)",
					"--builder-image",
					"$(inputs.params.BUILDER_IMAGE)",
					"--cache",
					"/cache",
					"--platform",
					"/platform",
					"--buildpack",
					"$(inputs.params.BUILDPACK)",
					"--platform-env",
					strings.Join(platformEnvSet.List(), ","),
				},
				VolumeMounts: cacheAndLayers,
			},
			{
				Name:    "detect",
				Image:   "$(inputs.params.BUILDER_IMAGE)",
				Command: []string{"/bin/bash"},
				Args: []string{
					"-c",
					`
if [[ -z "$(inputs.params.BUILDPACK)" ]]; then
  /lifecycle/detector \
    -app=/layers/source \
    -group=/layers/group.toml \
    -plan=/layers/plan.toml \
    -platform=/platform
else
  touch /layers/plan.toml
  echo -e "[[buildpacks]]\nid = \"$(inputs.params.BUILDPACK)\"\nversion = \"latest\"\n" > /layers/group.toml
fi
						`,
				},
				VolumeMounts: cacheAndLayers,
			},
			{
				Name:    "restore",
				Image:   "$(inputs.params.BUILDER_IMAGE)",
				Command: []string{"/lifecycle/restorer"},
				Args: []string{
					"-group=/layers/group.toml",
					"-layers=/layers",
					"-cache-dir=/cache",
				},
				VolumeMounts: cacheAndLayers,
			},
			{
				Name:    "build",
				Image:   "$(inputs.params.BUILDER_IMAGE)",
				Command: []string{"/lifecycle/builder"},
				Args: []string{
					"-app=/layers/source",
					"-layers=/layers",
					"-group=/layers/group.toml",
					"-plan=/layers/plan.toml",
					"-platform=/platform",
				},
				VolumeMounts: cacheAndLayers,
				Resources:    resources,
			},
			{
				Name:    "download-token",
				Image:   cfg.BuildTokenDownloadImage,
				Command: []string{"bash"},
				Args: []string{
					// TODO(b/169582594): This should likely reference a
					// container that was written in Go that is better
					// tested.
					"-c",
					`
cat << 'EOF' > /tmp/token.py
from urllib.parse import urlparse
import sys
import json


def extract_gcp_cr(u):
    o = urlparse("http://" + u)
    if o is None or o.hostname is None:
        return None

    # Exclude the subdomain.
    domain = '.'.join(o.hostname.split('.')[-2:])
    if domain == "gcr.io" or domain == "pkg.dev":
        return o.hostname
    else:
        return None


def write_token(cr, output_path, token):
    if cr is None:
        return

    data = {cr: "Bearer " + token}

    with open(output_path, 'w') as f:
        json.dump(data, f)


def main():
    # Check to see that we have 3 args and none are empty.
    if len(sys.argv) != 4 or not all(sys.argv[1:]):
        print("Usage: {name} [IMAGE_PATH] [OUTPUT] [TOKEN]".format(name = sys.argv[0]))
        sys.exit(1)
    image_path = sys.argv[1]
    output_path = sys.argv[2]
    token = sys.argv[3]
	$(joseph)
    write_token(extract_gcp_cr(image_path), output_path, token)


if __name__ == '__main__':
    main()
EOF

# This will retry a few times in case the gcloud command failed (i.e., WI token
# hasn't had time to propagate).
googleServiceAccount=$1
if [ "$googleServiceAccount" = "" ]; then
    exit 0
fi

# Retry for 2 minutes.
for i in $$(seq 1 24); do
    token="$$(gcloud auth application-default print-access-token)"
    if python3 /tmp/token.py $(inputs.params.DESTINATION_IMAGE) /workspace/gcloud.token "${token}"; then
        # Success
        exit 0
    else
        # Failure
        echo "failed to download token. Retrying..."
        sleep 5
    fi
done

# Never worked...
exit 1
`,
					"_",
					googleServiceAccount,
				},
				VolumeMounts: cacheAndLayers,
			},
			{
				Name:    "export",
				Image:   "$(inputs.params.BUILDER_IMAGE)",
				Command: []string{"bash"},
				Args: []string{
					"-c",
					`
set -eu

# CNB_REGISTRY_AUTH is used by /lifecycle/exporter
googleServiceAccount=$1
if [ "$googleServiceAccount" = "" ]; then
  echo "Workload Identity is not used"
else
  export CNB_REGISTRY_AUTH=$(cat /workspace/gcloud.token)
fi

# TODO: If https://github.com/buildpacks/lifecycle/issues/423 is resolved, then
# this can be replaced with /ko-app/build-helpers publish
export_image () {
  /lifecycle/exporter \
    -app=/layers/source \
    -layers=/layers \
    -group=/layers/group.toml \
    -image=$(inputs.params.RUN_IMAGE) \
    $(inputs.params.DESTINATION_IMAGE)
}

# This will retry a few times (2 minutes) in case exporting failed (i.e., WI
# token hasn't had time to propagate).
for i in $$(seq 1 24); do
	if export_image; then
        # Success
        exit 0
    else
        # Failure
        echo "failed to export image. Retrying..."
        sleep 5
    fi
done
`,
					"_",
					googleServiceAccount,
				},
				VolumeMounts: cacheAndLayers,
			},
			{
				Name:       "write results",
				WorkingDir: "/workspace",
				Command:    []string{"/ko-app/build-helpers"},
				Image:      cfg.BuildHelpersImage,
				Args: []string{
					"write-result",
					"$(inputs.params.DESTINATION_IMAGE)",
					"$(results.IMAGE.path)",
				},
			},
		},
		Volumes: []corev1.Volume{
			tektonutil.EmptyVolume("cache-dir"),
			tektonutil.EmptyVolume("layers-dir"),
			tektonutil.EmptyVolume("platform-dir"),
		},
	}
}
