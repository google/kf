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
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/compute/metadata"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/utils"
	"github.com/GoogleCloudPlatform/kf/samples/plugins/gke/pkg/commands"
)

func main() {
	projID, err := metadata.ProjectID()
	if err != nil {
		log.Fatalf("failed to fetch GCP project ID: %s", err)
	}

	cfg := utils.InBuildParseConfig()

	c := commands.PushCommand()
	c.Flags().Set("source-image", os.Getenv("SOURCE_IMAGE"))
	c.Flags().Set("container-registry", fmt.Sprintf("gcr.io/%s", projID))
	c.SetArgs(cfg.AllArgs)

	if err := c.Execute(); err != nil {
		log.Fatal(err)
	}
}
