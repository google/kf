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
	"flag"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/generator"
	"github.com/GoogleCloudPlatform/kf/samples/plugins/gke/pkg/commands"
)

func main() {
	log.SetFlags(0)

	var (
		containerRegistry string
		commandSetName    string
	)

	flag.StringVar(&containerRegistry, "container-registry", "", "the container registry the images are located (REQUIRED)")
	flag.StringVar(&commandSetName, "command-set-name", "kf", "the name of the generated CommandSet (defaults to 'kf')")
	flag.Parse()

	if containerRegistry == "" {
		log.Fatal(`--container-registry is required. Recommended --container-registry="$(gcloud config get-value project)"`)
	}

	generator.Register(commands.PushCommand, generator.WithRegisterUploadDir(true))
	generator.Convert(containerRegistry, commandSetName, os.Stdout)
}
