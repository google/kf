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

package commanddocgenerator

import (
	"bytes"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/russross/blackfriday/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"
)

// PrintFlags converts the flagSet to a GCP docs compatible
// format.
func PrintFlags(buf *bytes.Buffer, flagSet *pflag.FlagSet) {
	fmt.Fprintln(buf, "<dl>")
	defer fmt.Fprintln(buf, "</dl>")

	flagSet.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden || len(flag.Deprecated) != 0 {
			return
		}

		varname, usage := pflag.UnquoteUsage(flag)
		fmt.Fprint(buf, `<dt><code translate="no">`)
		if flag.Shorthand != "" {
			fmt.Fprintf(buf, "-%s, --%s", flag.Shorthand, flag.Name)
		} else {
			fmt.Fprintf(buf, "--%s", flag.Name)
		}
		if varname != "" {
			fmt.Fprintf(buf, `=<var translate="no">%s</var>`, varname)
		}

		fmt.Fprintln(buf, "</code></dt>")

		line := usage

		switch flag.DefValue {
		case "":
			// zero string value, don't print it
		case fmt.Sprintf("%v", 0):
			// zero numeric value, don't print it
		case fmt.Sprintf("%v", nil):
			// zero nil value, don't print it
		case fmt.Sprintf("%v", false):
			// zero bool value, don't print it
		case fmt.Sprintf("%v", []interface{}{}):
			// zero array value, don't print it
		default:
			if flag.Value.Type() == "string" {
				line += fmt.Sprintf(" (default %q)", flag.DefValue)
			} else {
				line += fmt.Sprintf(" (default %s)", flag.DefValue)
			}
		}
		// Convert usage text from Markdown to HTML,
		// this is needed because blackfriday lib converts markdown links to HTML hyper links correctly.
		lineBytes := []byte(line)
		formattedLine := blackfriday.Run(lineBytes, blackfriday.WithNoExtensions())
		fmt.Fprintln(buf, "<dd>"+string(formattedLine)+"</dd>")
	})
}

// TraverseCommands will invoke the given function for each command that is
// not hidden or only a help topic.
func TraverseCommands(cmd *cobra.Command, f func(cmd *cobra.Command) error) error {
	for _, c := range cmd.Commands() {
		if err := TraverseCommands(c, f); err != nil {
			return err
		}

		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := f(c); err != nil {
			return err
		}
	}

	return nil
}

type toc struct {
	Toc []tocEntry `json:"toc"`
}

type tocEntry struct {
	Title   string    `json:"title,omitempty"`
	Path    string    `json:"path,omitempty"`
	Section []section `json:"section,omitempty"`
}

type section struct {
	Title string `json:"title,omitempty"`
	Path  string `json:"path,omitempty"`
}

// GenerateBookYAML generates the _book.yaml.
// See https://developers.google.com/devsite/reference/metadata/book
func GenerateBookYAML(buf *bytes.Buffer, cmd *cobra.Command, docVersion string) {
	t := tocEntry{
		Title: "kf Command reference",
		Path:  fmt.Sprintf("/migrate/kf/docs/%s/cli/", docVersion),
	}

	TraverseCommands(cmd, func(cmd *cobra.Command) error {
		heritage := ListHeritage(cmd)

		title := strings.Join(heritage, " ")

		t.Section = append(t.Section, section{
			Title: title,
			Path:  path.Join(t.Path, strings.Join(heritage, "-")),
		})
		return nil
	})

	// Sort everything by their titles.
	sort.Slice(t.Section, func(i, j int) bool {
		return t.Section[i].Title < t.Section[j].Title
	})

	data, err := yaml.Marshal(toc{Toc: []tocEntry{t}})
	if err != nil {
		// This should never happen...
		panic(err)
	}
	buf.Write(data)
}

// ListHeritage returns a slice of names representing the heritage of a
// command.
func ListHeritage(cmd *cobra.Command) []string {
	h := []string{}

	// r has to be defined before being instantiated so it can be used
	// recursively.
	var r func(cmd *cobra.Command)
	r = func(cmd *cobra.Command) {
		if cmd.HasParent() {
			r(cmd.Parent())
		}

		if cmd.Use == "" {
			return
		}
		name := strings.Fields(cmd.Use)[0]

		h = append(h, name)
	}
	r(cmd)

	return h
}
