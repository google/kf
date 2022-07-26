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
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/kf/v2/pkg/dockerutil"
	"github.com/google/kf/v2/pkg/sourceimage"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewExtractCommand creates a command that will extract the contents of a
// container to a given location.
func NewExtractCommand() *cobra.Command {
	var (
		targetPath             string
		sourceImage            string
		sourcePackageName      string
		sourcePackageNamespace string
		sourcePath             string
		retryDuration          time.Duration
	)

	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract the contents of a container",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if sourceImage != "" && sourcePackageName != "" {
				log.Fatalf("--source-image and --source-package-name are mutually exclusive")
			}
			if sourceImage == "" && sourcePackageName == "" {
				log.Fatalf("one of --source-image or --source-package-name are required")
			}
			if sourcePackageName != "" && sourcePackageNamespace == "" {
				log.Fatalf("--source-package-namespace must be provided with --source-package-name")
			}

			if sourceImage != "" {
				return extractSourceImageWithRetry(
					retryDuration,
					sourceImage,
					sourcePath,
					targetPath)
			}
			return extractSourcePackage(
				cmd.Context(),
				sourcePackageNamespace,
				sourcePackageName,
				sourcePath,
				targetPath)
		},
	}

	cmd.Flags().StringVar(&targetPath, "output-dir", "/workspace/source", "the workspace for the app")
	cmd.Flags().StringVar(&sourceImage, "source-image", "", "the image containing source to extract")
	cmd.Flags().StringVar(&sourcePackageNamespace, "source-package-namespace", "", "the namespace of the source package")
	cmd.Flags().StringVar(&sourcePackageName, "source-package-name", "", "the name of the source package")
	cmd.Flags().StringVar(&sourcePath, "source-path", sourceimage.DefaultSourcePath, "the directory to extract from the source")
	cmd.Flags().DurationVar(&retryDuration, "retry-duration", 2*time.Minute, "the amount of time to retry due to 403s (Workload Identity)")

	return cmd
}

func createOutputDir(path string) error {
	log.Printf("Creating output directory: %s", path)
	if err := os.MkdirAll(path, 0755); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func getClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

const (
	sourcePackageSubresourceApiFormat = "apis/upload.kf.dev/v1alpha1/proxy/namespaces/%s/%s"

	// extractSourcePackageTimeoutInMinutes is the timeout value for extracting source package from subresource api-server.
	// This value is a wild guess.
	// The default value is under a minute.
	extractSourcePackageTimeoutInMinutes = 15
)

// extractSourcePackage extracts the source image by speaking the subresource
// api server to get the image associated with a SourcePackage then extracing
// the tar. This method avoids issues with Workload Identity availability on
// startup.
func extractSourcePackage(ctx context.Context, sourcePackageNamespace, sourcePackageName, sourcePath, targetPath string) error {
	if err := createOutputDir(targetPath); err != nil {
		return err
	}
	client, err := getClient()
	if err != nil {
		return err
	}
	path := fmt.Sprintf(sourcePackageSubresourceApiFormat, sourcePackageNamespace, sourcePackageName)
	log.Printf("Getting sourcepackage via subresource api at path: %s\n", path)

	var code int
	result := client.Discovery().
		RESTClient().
		Get().
		RequestURI(path).
		Timeout(extractSourcePackageTimeoutInMinutes * time.Minute).
		Do(ctx)
	result.StatusCode(&code)
	bs, err := result.Raw()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get sourcepackage (http code %d): %s", code, string(bs)))
	}

	r := tar.NewReader(bytes.NewReader(bs))
	err = sourceimage.ExtractTar(targetPath, sourcePath, r)
	if err != nil {
		return errors.Wrap(err, "failed to extract tar")
	}
	return nil
}

// extractSourceImage extracts the source image by speaking directly with the
// container registry, downloading the image, and extracting the tar.
func extractSourceImage(sourceImage, sourcePath, targetPath string) error {
	if err := createOutputDir(targetPath); err != nil {
		return err
	}
	log.Printf("Fetching image: %s", sourceImage)
	dockerutil.DescribeDefaultConfig(os.Stdout)
	dockerutil.DescribeWorkloadIdentity(os.Stdout)
	imageRef, err := name.ParseReference(sourceImage, name.WeakValidation)
	if err != nil {
		return err
	}

	image, err := remote.Image(imageRef, dockerutil.GetAuthKeyChain())
	if err != nil {
		return err
	}

	log.Printf("Extracting contents from: %s to %s\n", sourcePath, targetPath)
	return sourceimage.ExtractImage(targetPath, sourcePath, image)
}

func extractSourceImageWithRetry(retryDuration time.Duration, sourceImage, sourcePath, targetPath string) error {
	start := time.Now()

	// TODO: This is basic retry (without exponential backoff). However we can
	// later import k8s.io/client-go/util/retry when we update to v0.16.9 and
	// use OnError().  We can't use github.com/googleapis/gax-go/v2 as we do
	// in the space reconciler as that expects gPRC codes.
	for ; ; time.Sleep(5 * time.Second) {
		err := extractSourceImage(sourceImage, sourcePath, targetPath)
		switch {
		case err == nil: // success
			log.Println("Success")
			return nil

		case time.Since(start) > retryDuration: // timed out
			return err

		default:
			if me, ok := err.(*metadata.Error); ok && me.Code == http.StatusForbidden {
				log.Printf("failed with a 403 (retrying...): %v\n", err)
				continue
			}
			return err
		}
	}
}
