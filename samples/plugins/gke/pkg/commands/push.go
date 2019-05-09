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

package commands

import (
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/spf13/cobra"
)

func PushCommand() *cobra.Command {
	var p config.KfParams
	c := commands.InjectPush(&p)

	c.Use = "push APP_NAME [flags]"
	c.Example = `
 push myapp
 push myapp --buildpack my.special.buildpack # Discover via kf buildpacks
 push myapp --env FOO=bar --env BAZ=foo
`

	c.PersistentFlags().StringVar(&p.Namespace, "namespace", "default", "kubernetes namespace")
	c.Flags().MarkHidden("container-registry")
	c.Flags().MarkHidden("service-account")
	return c
}
