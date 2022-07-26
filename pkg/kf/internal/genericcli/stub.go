// Copyright 2020 Google LLC
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

package genericcli

import (
	"fmt"

	"github.com/google/kf/v2/third_party/k8s.io/kubectl/pkg/util/templates"
	"github.com/spf13/cobra"
)

// NewStubCommand creates a stub command.
func NewStubCommand(name string, short string, alt string, example string, opts ...StubOption) *cobra.Command {
	options := StubOptions(opts)
	long := fmt.Sprintf("kf %s is currently unsupported.\n%s\n", name, templates.LongDesc(alt))
	return &cobra.Command{
		Use:                name,
		Short:              short,
		Long:               long,
		Aliases:            options.Aliases(),
		Args:               cobra.ArbitraryArgs,
		Example:            example,
		DisableFlagParsing: true,
		SilenceUsage:       true,

		// Stubs MUST NOT have a Run or RunE
	}
}
