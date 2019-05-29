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
	"fmt"
	"os"
	"path/filepath"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/kf"
	"github.com/spf13/cobra"
)

// NewUploadBuildpacks creates a UploadBuildpacks command.
func NewUploadBuildpacks(p *config.KfParams, c buildpacks.BuilderCreator, u buildpacks.BuildTemplateUploader) *cobra.Command {
	var (
		containerRegistry string
		path              string
	)
	var uploadBuildpacksCmd = &cobra.Command{
		Use:   "upload-buildpacks",
		Short: "Create and upload a new buildpacks builder. This is used to set the available buildpacks that are used while pushing an app.",
		Args:  cobra.ExactArgs(0),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if p.Namespace != "default" && p.Namespace != "" {
				fmt.Fprintf(cmd.OutOrStderr(), "NOTE: Buildpacks are global and are available to all spaces.")
			}

			if path != "" {
				var err error
				path, err = filepath.Abs(path)
				if err != nil {
					return err
				}
			} else {
				var err error
				path, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			image, err := c.Create(path, containerRegistry)
			if err != nil {
				cmd.SilenceUsage = !kf.ConfigError(err)
				return err
			}

			if err := u.UploadBuildTemplate(image); err != nil {
				cmd.SilenceUsage = !kf.ConfigError(err)
				return err
			}

			return nil
		},
	}

	uploadBuildpacksCmd.Flags().StringVar(
		&path,
		"path",
		"",
		"The path the source code lives. Defaults to current directory.",
	)

	uploadBuildpacksCmd.Flags().StringVar(
		&containerRegistry,
		"container-registry",
		"",
		"The container registry to push the resulting container.",
	)

	return uploadBuildpacksCmd
}
