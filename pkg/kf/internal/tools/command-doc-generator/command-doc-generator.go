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

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/kf/pkg/kf/commands"
	"github.com/spf13/cobra/doc"
)

const (
	fmTemplate = `---
title: "%s"
slug: %s
url: %s
---
`

	indexTemplate = `
---
title: "Commands"
linkTitle: "Commands"
weight: 20
description: "Reference for the kf CLI"
---
`

	prefix = "/docs/general-info/kf-cli/commands/"
)

func main() {
	kfOnly := flag.Bool("kf-only", false, "when set, the output will write the kf command. This is useful to generate the _index.md")
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalf("usage: %s <OUTPUT PATH>", os.Args[0])
	}
	outputPath := flag.Args()[0]

	kf := commands.NewKfCommand()
	filePrepender := func(filename string) string {
		name := filepath.Base(filename)
		title := strings.ReplaceAll(strings.TrimSuffix(name, path.Ext(name)), "_", " ")
		base := strings.ReplaceAll(strings.TrimSuffix(name, path.Ext(name)), "_", "-")
		url := prefix + strings.ToLower(base) + "/"
		return fmt.Sprintf(fmTemplate, title, base, url)
	}
	linkHandler := func(name string) string {
		base := strings.ReplaceAll(strings.TrimSuffix(name, path.Ext(name)), "_", "-")
		return prefix + strings.ToLower(base) + "/"
	}

	if *kfOnly {
		buf := &bytes.Buffer{}
		fmt.Fprintln(buf, indexTemplate)

		if err := doc.GenMarkdownCustom(kf, buf, linkHandler); err != nil {
			log.Fatal(err)
		}

		if err := ioutil.WriteFile(outputPath, buf.Bytes(), 0666); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err := doc.GenMarkdownTreeCustom(kf, outputPath, filePrepender, linkHandler); err != nil {
		log.Fatal(err)
	}
}
