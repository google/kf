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

//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/kf/v2/pkg/kf/doctor/troubleshooter"
	runbookgenerator "github.com/google/kf/v2/pkg/kf/internal/tools/runbook-generator"
)

func main() {
	flag.Parse()

	if len(flag.Args()) != 2 {
		log.Fatalf("usage: %s OUTPUT_PATH DOC_VERSION", os.Args[0])
	}

	args := flag.Args()
	outputPath := args[0]
	docVersion := args[1]

	if err := genRunbookMarkdown(outputPath, docVersion); err != nil {
		log.Fatalf("error generating markdown: %v", err)
	}
}

func genRunbookMarkdown(outputPath, docVersion string) error {
	if err := os.MkdirAll(outputPath, 0700); err != nil {
		return err
	}

	for _, component := range troubleshooter.CustomResourceComponents() {
		buf := &bytes.Buffer{}
		runbookgenerator.GenTroubleshooterRunbook(context.Background(), buf, component, docVersion)
		fileName := fmt.Sprintf("troubleshoot_%s.md", strings.ToLower(component.Type.FriendlyName()))
		path := filepath.Join(outputPath, fileName)

		log.Println("Generating", path)
		if err := ioutil.WriteFile(path, buf.Bytes(), 0600); err != nil {
			return err
		}
	}

	return nil
}
