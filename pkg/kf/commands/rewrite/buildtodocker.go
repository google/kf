package rewrite

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	apiconfig "github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/manifest"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

//go:embed resources/Dockerfile
var rawDockerfileTemplate []byte

//go:embed resources/config-defaults.yaml
var configDefaultsYaml []byte

//go:embed resources/launcher
var launcherSource embed.FS

func parseConfigDefaults() (*apiconfig.DefaultsConfig, error) {
	configDefaultsConfigmap := new(corev1.ConfigMap)
	if err := yaml.Unmarshal(configDefaultsYaml, configDefaultsConfigmap); err != nil {
		return nil, err
	}

	return apiconfig.NewDefaultsConfigFromConfigMap(configDefaultsConfigmap)
}

func extractLauncherSource(rootDirectory string) error {
	return fs.WalkDir(launcherSource, "", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		outputPath := filepath.Join(".", "launchershim", filepath.FromSlash(path))
		switch {
		case d.IsDir():
			if err := os.MkdirAll(outputPath, d.Type()); err != nil {
				return err
			}
		default:
			fd, err := os.Open(outputPath)
			if err != nil {
				return err
			}
			defer fd.Close()

			sfd, err := launcherSource.Open(d.Name())
			if err != nil {
				return err
			}
			defer sfd.Close()

			if _, err := io.Copy(fd, sfd); err != nil {
				return err
			}
		}
		return nil
	})
}

// NewBuildToDocker rewrites the apps in a Kf manifest to Dockerfiles.
func NewBuildToDocker(cfg *config.KfParams) *cobra.Command {
	cmd := &cobra.Command{
		Hidden:  true,
		Use:     "build-to-docker",
		Short:   "Create a dockerfile encapsulating the build for the app in the current directory.",
		Example: `kf build-to-docker`,
		Args:    cobra.ExactArgs(0),
		Long:    `Produces a Dockerfile per-app.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			appManifest, err := manifest.CheckForManifest(context.Background(), nil)
			if err != nil {
				return err
			}

			if appManifest == nil {
				appManifest, err = manifest.New("default")
				if err != nil {
					return err
				}
			}

			// Parse common config
			cd, err := parseConfigDefaults()
			if err != nil {
				return fmt.Errorf("couldn't parse config-defaults: %w", err)
			}

			buildConfig := v1alpha1.SpaceStatusBuildConfig{
				BuildpacksV2: cd.SpaceBuildpacksV2,
				StacksV2:     cd.SpaceStacksV2,
				StacksV3:     cd.SpaceStacksV3,
			}

			// Parse template
			dockerTemplate, err := template.New("Dockerfile").Parse(string(rawDockerfileTemplate))
			if err != nil {
				return fmt.Errorf("couldn't parse Dockerfile: %w", err)
			}

			// Once
			// 		Move a kfignore/cfignore to a dockerignore
			//		Create the launcher
			fmt.Fprintf(cmd.OutOrStdout(), "Extracting launcher code\n")
			extractLauncherSource(".")

			fmt.Fprintf(cmd.OutOrStdout(), "Found %d application(s)\n", len(appManifest.Applications))
			for _, app := range appManifest.Applications {
				fmt.Fprintf(cmd.OutOrStdout(), "Updating app %q\n", app.Name)

				buildSpecBuilder, _, err := app.DetectBuildType(buildConfig)
				if err != nil {
					return err
				}
				buildSpec, err := buildSpecBuilder("source-image://invalid-value")
				if err != nil {
					return err
				}
				if buildSpec.Name != v1alpha1.BuildpackV2BuildTaskName {
					fmt.Fprintf(cmd.OutOrStdout(), "Not a v2 built app: %q", buildSpec.Name)
					continue
				}
				params := make(map[string]interface{})
				for _, param := range buildSpec.Params {
					params[param.Name] = param.Value
				}
				localSource := "."
				if app.Path != "" {
					localSource = app.Path
				}
				params["LOCAL_SOURCE"] = localSource

				if err := dockerTemplate.Execute(cmd.OutOrStdout(), params); err != nil {
					return err
				}
			}

			// For every app:
			// 		Create a Dockerfile.<appname>
			//

			return nil
		},
	}
	return cmd
}
