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

package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewKfPrintFlags creates a new KfPrintFlags.
func NewKfPrintFlags() *KfPrintFlags {
	return &KfPrintFlags{
		PrintFlags: genericclioptions.NewPrintFlags(""),
	}
}

// KfPrintFlags adapts changes that have to be made to Kubernetes'
// PrintFlags to work with Kf's documentation and style guidelines.
type KfPrintFlags struct {
	*genericclioptions.PrintFlags
}

// AddFlags overrides genericclioptions.PrintFlags::AddFlags.
func (f *KfPrintFlags) AddFlags(cmd *cobra.Command) {
	f.PrintFlags.AddFlags(cmd)

	// Override output format to be sorted so our generated documents are deterministic
	allowedFormats := f.PrintFlags.AllowedFormats()
	sort.Strings(allowedFormats)
	cmd.Flag("output").Usage = fmt.Sprintf("Output format. One of: %s.", strings.Join(allowedFormats, "|"))

	// Override template description to match our other docs.
	cmd.Flag("template").Usage = "Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is [golang templates](http://golang.org/pkg/text/template/#pkg-overview)."
}
