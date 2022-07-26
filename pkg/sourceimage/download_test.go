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

package sourceimage

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck/v1beta1"
)

func makeSourcePackage(imageRef string, ready bool) *v1alpha1.SourcePackage {
	sp := v1alpha1.SourcePackage{
		Status: v1alpha1.SourcePackageStatus{
			Image: imageRef,
		},
	}
	if ready {
		sp.Status = v1alpha1.SourcePackageStatus{
			Image: imageRef,
			Status: v1beta1.Status{
				Conditions: v1beta1.Conditions{
					apis.Condition{
						Type:   apis.ConditionReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}
	}
	return &sp
}

func makeTarImage(t *testing.T, filename, body string) v1.Image {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	tw.WriteHeader(&tar.Header{
		Name: filename,
		Mode: 0755,
		Size: int64(len(body)),
	})
	tw.Write([]byte(body))
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if dataLayer, err := tarball.LayerFromReader(&buf); err != nil {
		t.Fatal(err)
	} else if img, err := mutate.AppendLayers(empty.Image, dataLayer); err != nil {
		t.Fatal(err)
	} else {
		return img
	}
	return nil
}

func assertTar(t *testing.T, r io.ReadCloser, filename, body string) {
	tr := tar.NewReader(r)
	header, err := tr.Next()
	testutil.AssertNil(t, "err", err)
	bs, _ := ioutil.ReadAll(tr)
	testutil.AssertEqual(t, "header", filename, header.Name)
	testutil.AssertEqual(t, "body", body, string(bs))
}

func makeRef(registryUrl string) (name.Reference, error) {
	ref := strings.TrimPrefix(registryUrl, "http://") + "/source-package"
	return name.NewTag(ref)
}

func TestDownload(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		sourcePackageFn func(ref name.Reference) *v1alpha1.SourcePackage
		image           v1.Image
		expectExtract   []string
		expectErr       string
	}{
		"uploaded source package": {
			sourcePackageFn: func(ref name.Reference) *v1alpha1.SourcePackage {
				return makeSourcePackage(ref.String(), true)
			},
			image:         makeTarImage(t, "foo", "foo bar"),
			expectExtract: []string{"foo", "foo bar"},
		},
		"uploaded but not available": {
			sourcePackageFn: func(ref name.Reference) *v1alpha1.SourcePackage {
				return makeSourcePackage(ref.String(), true)
			},
			expectErr: "failed to get image",
		},
		"not uploaded source package": {
			sourcePackageFn: func(ref name.Reference) *v1alpha1.SourcePackage {
				return makeSourcePackage(ref.String(), false)
			},
			expectErr: "source package not yet uploaded",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fp := NewFakeSourcePackageLister(ctrl)
			fnp := NewFakeSourcePackageNamespaceLister(ctrl)

			reg := registry.New(registry.Logger(log.New(ioutil.Discard, "", 0)))
			s := httptest.NewServer(reg)
			defer s.Close()

			ref, err := makeRef(s.URL)
			testutil.AssertNil(t, "tag err", err)

			namespace := "foo"
			sourcePackage := "source-package"

			fp.EXPECT().SourcePackages(namespace).Return(fnp)
			fnp.EXPECT().Get(sourcePackage).Return(tc.sourcePackageFn(ref), nil)

			if tc.image != nil {
				remote.Write(ref, tc.image)
			}

			r, err := Download(fp, namespace, sourcePackage)
			defer func() {
				if r != nil {
					r.Close()
				}
			}()

			if tc.expectExtract != nil {
				filename, body := tc.expectExtract[0], tc.expectExtract[1]
				assertTar(t, r, filename, body)
				testutil.AssertNil(t, "Download err", err)
			} else {
				testutil.AssertErrorContainsAll(t, err, []string{tc.expectErr})
			}
		})
	}
}
