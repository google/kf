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

//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/kf/v2/pkg/kf/commands"
	commanddocgenerator "github.com/google/kf/v2/pkg/kf/internal/tools/command-doc-generator"
	"github.com/spf13/cobra"
)

const (
	fmTemplate = `---
title: %q
weight: 100
description: %q
---
`
)

func main() {
	flag.Parse()

	if len(flag.Args()) != 2 {
		log.Fatalf("usage: %s <OUTPUT PATH> <DOC VERSION>", os.Args[0])
	}
	outputPath := flag.Args()[0]
	docVersion := flag.Args()[1]

	kf := commands.NewRawKfCommand()

	if err := genMarkdownTree(kf, outputPath, docVersion); err != nil {
		log.Fatal(err)
	}
}

// genCommandMarkdown creates markdown output for the command that works
// for the Docsy Hugo template.
func genCommandMarkdown(cmd *cobra.Command, w io.Writer) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	short := cmd.Short
	long := cmd.Long
	if len(long) == 0 {
		long = short
	}

	// Set the header
	if _, err := fmt.Fprintf(w, fmTemplate, cmd.CommandPath(), short); err != nil {
		return err
	}

	// Structure based on gcloud's docs:
	// https://cloud.google.com/sdk/gcloud/reference/beta/help

	buf.WriteString("### Name\n\n")
	buf.WriteString(fmt.Sprintf(`<code translate="no">%s</code> - %s`, name, short))
	buf.WriteString("\n\n")

	if cmd.Runnable() {
		buf.WriteString("### Synopsis\n\n")
		buf.WriteString(fmt.Sprintf(`<pre translate="no">%s</pre>`, cmd.UseLine()))
		buf.WriteString("\n\n")
	}

	if short != long {
		buf.WriteString("### Description\n\n")
		buf.WriteString(heredoc.Doc(long) + "\n\n")
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("### Examples\n\n")
		buf.WriteString(`<pre translate="no">`)
		buf.WriteString("\n")
		buf.WriteString(html.EscapeString(heredoc.Doc(cmd.Example)))
		buf.WriteString("</pre>\n\n")
	}

	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("### Flags\n\n")
		commanddocgenerator.PrintFlags(buf, flags)
		buf.WriteString("\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("### Inherited flags\n\n")
		buf.WriteString("These flags are inherited from parent commands.\n\n")
		commanddocgenerator.PrintFlags(buf, parentFlags)
		buf.WriteString("\n\n")
	}

	_, err := buf.WriteTo(w)
	return err
}

func genMarkdownTree(cmd *cobra.Command, dir, docVersion string) error {
	commanddocgenerator.TraverseCommands(cmd, func(cmd *cobra.Command) error {
		return genMarkdownTree(cmd, dir, docVersion)
	})

	// TODO: figure out if this is the root command and use index.html instead.
	basename := strings.Replace(cmd.CommandPath(), " ", "-", -1)
	basename = strings.Replace(basename, "_", "-", -1)
	filename := filepath.Join(dir, basename+".md")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := genCommandMarkdown(cmd, f); err != nil {
		return err
	}

	return nil
}
