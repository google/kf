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
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestUpstreamPrintFlags(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().SortFlags = true

	printFlags := genericclioptions.NewPrintFlags("")
	printFlags.AddFlags(cmd)

	flagNames := sets.NewString()
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flagNames.Insert(flag.Name)
	})

	// XXX: If this test fails, it means the upstream library has changed.
	// Validate that new flags don't need fixes before adjusting.
	testutil.AssertEqual(t,
		"flags",
		[]string{"allow-missing-template-keys", "output", "show-managed-fields", "template"},
		flagNames.List(),
	)
}

func ExampleNewKfPrintFlags() {
	cmd := &cobra.Command{}
	cmd.Flags().SortFlags = true

	printFlags := NewKfPrintFlags()
	printFlags.AddFlags(cmd)

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		fmt.Println(flag.Name, " ", flag.Usage)
	})

	// XXX: Kf expects determinstic output for these flags.

	// Output: allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats.
	// output   Output format. One of: go-template|go-template-file|json|jsonpath|jsonpath-as-json|jsonpath-file|name|template|templatefile|yaml.
	// show-managed-fields   If true, keep the managedFields when printing objects in JSON or YAML format.
	// template   Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is [golang templates](http://golang.org/pkg/text/template/#pkg-overview).
}
