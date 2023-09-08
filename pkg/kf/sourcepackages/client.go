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

package sourcepackages

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	cv1alpha1 "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/typed/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/sourceimage"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
)

// ClientExtension holds additional functions that should be exposed by
// Client.
type ClientExtension interface {
	// UploadSourcePath packages up (normally via tar) the given source path,
	// creates a SourcePackage for it, and finally POSTs it to the
	// upload.kf.dev endpoint.
	UploadSourcePath(ctx context.Context, sourcePath string, app *v1alpha1.App) error
}

// Poster is a function used to write data to the API server.
type Poster func(
	ctx context.Context,
	requestURI string,
	bodyFileName string,
) error

type posterErr struct {
	err  error
	body string
}

var _ error = (*posterErr)(nil)

// Error implements error interface.
func (e *posterErr) Error() string {
	return fmt.Sprintf("%v: (body=%s)", e.err, e.body)
}

// ExtractRequestError will return the underlying request error if there is
// one. This is useful if the request fails and the error has information that
// would make it actionable.
func ExtractRequestError(err error) error {
	e, ok := err.(*posterErr)
	if !ok {
		// not a posterErr, just return it.
		return err
	}

	return e.err
}

// NewPoster returns a new Poster function from a Kubernetes Interface.
func NewPoster(ki kubernetes.Interface) Poster {
	return func(
		ctx context.Context,
		requestURI string,
		bodyFileName string,
	) error {
		req := ki.CoreV1().RESTClient().Post()
		if bodyData, err := req.RequestURI(requestURI).Body(bodyFileName).DoRaw(ctx); err != nil {
			return &posterErr{err: err, body: string(bodyData)}
		}

		return nil
	}
}

type sourcePackagesClient struct {
	coreClient
	poster Poster
}

// NewClient creates a new application client.
func NewClient(
	kclient cv1alpha1.SourcePackagesGetter,
	poster Poster,
) Client {
	return &sourcePackagesClient{
		coreClient: coreClient{
			kclient: kclient,
		},
		poster: poster,
	}
}

// UploadSourcePath implements ClientExtension.
func (p *sourcePackagesClient) UploadSourcePath(
	ctx context.Context,
	sourcePath string,
	app *v1alpha1.App,
) error {
	// Package up the given SourcePath and get its checksum.
	cleanup, tarFile, cz, err := PackageSourcePath(sourcePath)

	// Always invoke the cleanup function.
	defer cleanup()

	if err != nil {
		return fmt.Errorf("failed to package up source directory: %v", err)
	}

	// Get the size of the tar file.
	fi, err := tarFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info for tar file: %v", err)
	}

	spName := app.Spec.Build.Spec.SourcePackage.Name

	if _, err := p.Create(
		ctx,
		app.Namespace,
		&v1alpha1.SourcePackage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      spName,
				Namespace: app.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*kmeta.NewControllerRef(app),
				},
			},
			Spec: v1alpha1.SourcePackageSpec{
				Size: uint64(fi.Size()),
				Checksum: v1alpha1.SourcePackageChecksum{
					Type:  v1alpha1.PackageChecksumSHA256Type,
					Value: cz,
				},
			},
		},
	); err != nil {
		return fmt.Errorf("failed to create SourcePackage: %v", err)
	}

	// Upload the data.
	logger := logging.FromContext(ctx).With(zap.Namespace("source upload"))
	logger.Infof("Uploading directory %s (size=%0.2fKiB)", sourcePath, float64(fi.Size()/1024.0))
	uploadStart := time.Now()
	requestURI := fmt.Sprintf(
		"/apis/upload.kf.dev/v1alpha1/proxy/namespaces/%s/%s",
		app.Namespace,
		spName,
	)

	err = p.poster(ctx, requestURI, tarFile.Name())

	if err != nil {
		return fmt.Errorf("failed to upload source directory: %v", err)
	}

	logger.Infof("Successfully uploaded source in %0.2f seconds", time.Since(uploadStart).Seconds())

	// Success!
	return nil
}

// PackageSourcePath tars up the given path, takes the checksum. The cleanup
// function should ALWAYS be invoked, even if there is an error.
func PackageSourcePath(sourcePath string) (
	cleanup func(),
	tarFile *os.File,
	checksum string,
	err error,
) {
	cleanup = func() {
		// NOP.
		// This will be replaced later in the function when there is something
		// to do.
	}

	// PackageSourceTar expects absolute paths.
	sourcePathAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		return cleanup, nil, "", err
	}

	// Tar up the path.
	fileFilter, err := sourceimage.BuildIgnoreFilter(sourcePath)
	if err != nil {
		return cleanup, nil, "", err
	}

	tmpFile, err := ioutil.TempFile("", "kf-source")
	if err != nil {
		return cleanup, nil, "", err
	}

	cleanup = func() {
		// XXX: This is usiing the generic logger. Ideally we would have a
		// logger saved on a context so that we can use the configured STDERR,
		// however the CLI doesn't do that type of thing yet.
		if err := tmpFile.Close(); err != nil {
			log.Printf("failed to close temp file %s: %v", tmpFile.Name(), err)
		}
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Printf("failed to delete temp file %s: %v", tmpFile.Name(), err)
		}
	}

	h := sha256.New()

	if err := sourceimage.PackageSourceTar(
		io.MultiWriter(h, tmpFile),
		sourcePathAbs,
		fileFilter,
	); err != nil {
		return cleanup, nil, "", err
	}

	// Get the sha256 of the resultiing file.
	cz := h.Sum(nil)

	return cleanup, tmpFile, hex.EncodeToString(cz[:]), nil
}
