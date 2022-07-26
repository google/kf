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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const (
	requestURIFormat = "apis/upload.kf.dev/v1alpha1/proxy/namespaces/%s/%s/image"

	// publishImageTimeoutInMinutes is the timeout value for publishing images to the subresource api-server.
	// This value is a wild guess.
	// The default value is under a minute.
	publishImageTimeoutInMinutes = 15
)

func NewPublishCommand() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:     "publish TAR_PATH NAMESPACE BUILD_NAME",
		Example: `publish /tmp/image.tar my-space build-0`,
		Long: `
		publish uploads the given .tar file to subresource apiserver endpoint at 
		(apis/upload.kf.dev/v1alpha1/proxy/namespaces/<SPACE_NAME>/<BUILD_NAME>/image).

		The endpoint packages the .tar file as a container image, and pushes the image 
		to the container registry specified at the Space. The image URL is determined by 
		the container registry URL and the Build name.
		`,
		Short: "Publish a container image with the given tar path, namespace and build.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			tarPath := args[0]
			buildNamespace := args[1]
			buildName := args[2]

			ctx := cmd.Context()
			client, err := getClient()
			if err != nil {
				return err
			}
			requestURI := fmt.Sprintf(requestURIFormat, buildNamespace, buildName)
			bodyData, err := client.Discovery().
				RESTClient().
				Post().
				RequestURI(requestURI).
				Body(tarPath).
				Timeout(publishImageTimeoutInMinutes * time.Minute).
				DoRaw(ctx)
			if err != nil {
				return fmt.Errorf("failed to post data: %v: %s", err, bodyData)
			}

			var m map[string]string
			if err := json.Unmarshal(bodyData, &m); err != nil {
				return fmt.Errorf("failed to decode data: %v: %s", err, bodyData)
			}

			digest := m["digest"]
			if outputFile == "" {
				fmt.Fprint(cmd.OutOrStdout(), digest)
			} else {
				if err := ioutil.WriteFile(outputFile, []byte(digest), os.ModePerm); err != nil {
					return fmt.Errorf("failed to output data: %v", err)
				}
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&outputFile, "output", "", "the file path to write the output. If empty, it will write to stdout.")

	return cmd
}
