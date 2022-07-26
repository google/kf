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

package commands

import (
	"fmt"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	thirdparty "github.com/google/kf/v2/third_party"
	"github.com/spf13/cobra"
)

// NewThirdPartyLicensesCommand creates a command that prints third party
// license info.
func NewThirdPartyLicensesCommand() *cobra.Command {
	return &cobra.Command{
		Annotations: map[string]string{
			config.SkipVersionCheckAnnotation: "",
		},
		Use:   "third-party-licenses",
		Short: "Print third party license information.",
		Long: fmt.Sprintf(`
Kf depends on the following third-party components that come with their own
licenses:

<pre>
%s
</pre>
`, thirdparty.ThirdPartyLicenses),
		Example:      "kf third-party-licenses",
		SilenceUsage: true,
	}
}
