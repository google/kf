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
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/kf/v2/pkg/kf/commands"
	commanddocgenerator "github.com/google/kf/v2/pkg/kf/internal/tools/command-doc-generator"
	"github.com/spf13/cobra"
)

const (
	fmTemplate = `{%% extends "migrate/kf/docs/%s/_base.html" %%}
{%% block page_title %%} %s {%% endblock %%}

{%% block body %%}

`
	footer = `{% endblock %}`
)

func main() {
	kfOnly := flag.Bool("kf-only", false, "when set, the output will write the kf command. This is useful to generate the index.md")
	bookOnly := flag.Bool("book-only", false, "when set, the output will write the book.yaml file only.")
	flag.Parse()

	if len(flag.Args()) != 2 {
		log.Fatalf("usage: %s <OUTPUT PATH> <DOC VERSION>", os.Args[0])
	}
	outputPath := flag.Args()[0]
	docVersion := flag.Args()[1]

	kf := commands.NewRawKfCommand()

	if *kfOnly {
		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, fmTemplate, docVersion, "kf Command Reference")

		if err := genCommandMarkdown(kf, buf); err != nil {
			log.Fatal(err)
		}

		fmt.Fprintln(buf, footer)

		// Remove the version and documentation headers
		newBuf := &bytes.Buffer{}
		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			switch {
			case strings.HasPrefix(scanner.Text(), "Kf CLI Version:"), strings.HasPrefix(scanner.Text(), "{{kf_product_name_short}} CLI Version:"):
				continue
			case strings.HasPrefix(scanner.Text(), "Documentation:"):
				continue
			default:
				fmt.Fprintln(newBuf, scanner.Text())
			}
		}
		if err := ioutil.WriteFile(outputPath, newBuf.Bytes(), 0666); err != nil {
			log.Fatal(err)
		}

		return
	}

	if *bookOnly {
		buf := &bytes.Buffer{}
		commanddocgenerator.GenerateBookYAML(buf, kf, docVersion)

		if err := ioutil.WriteFile(outputPath, buf.Bytes(), 0666); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := genMarkdownTree(kf, outputPath, docVersion); err != nil {
		log.Fatal(err)
	}
}

// genCommandMarkdown creates markdown output for the command in Google Cloud
// style.
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

	match := regexp.MustCompile(`\bKf\b`)
	var newBuf bytes.Buffer
	newBuf.WriteString(match.ReplaceAllString(buf.String(), "{{kf_product_name_short}}"))

	_, err := newBuf.WriteTo(w)
	return err
}

func genMarkdownTree(cmd *cobra.Command, dir, docVersion string) error {
	commanddocgenerator.TraverseCommands(cmd, func(cmd *cobra.Command) error {
		return genMarkdownTree(cmd, dir, docVersion)
	})

	// TODO: figure out if this is the root command and use index.html instead.
	basename := strings.Replace(cmd.CommandPath(), " ", "-", -1) + ".md"
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Add the Google Cloud required header
	if _, err := fmt.Fprintf(f, fmTemplate, docVersion, cmd.CommandPath()); err != nil {
		return err
	}

	// Generate the command markdown
	if err := genCommandMarkdown(cmd, f); err != nil {
		return err
	}

	// Add the Google Cloud required footer to finish the template
	if _, err := fmt.Fprintln(f, footer); err != nil {
		return err
	}

	return nil
}
