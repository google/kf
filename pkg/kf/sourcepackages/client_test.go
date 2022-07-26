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
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	fake "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestPosterErr(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		err       error
		want      error
		errString string
	}{
		{
			name:      "normal error",
			err:       errors.New("some-error"),
			want:      errors.New("some-error"),
			errString: "some-error",
		},
		{
			name:      "nil error",
			err:       nil,
			want:      nil,
			errString: "<nil>",
		},
		{
			name:      "posterErr error",
			err:       &posterErr{err: errors.New("some-error"), body: "some-body"},
			want:      errors.New("some-error"),
			errString: "some-error: (body=some-body)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractRequestError(tc.err)
			testutil.AssertErrorsEqual(t, tc.want, got)
			testutil.AssertEqual(t, "errString", tc.errString, fmt.Sprint(tc.err))
		})
	}
}

func TestClient_UploadSourcePath(t *testing.T) {
	t.Parallel()

	type fakes struct {
		cs     *fake.Clientset
		poster Poster
	}

	app := &v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "some-namespace",
		},
		Spec: v1alpha1.AppSpec{
			Build: v1alpha1.AppSpecBuild{
				Spec: &v1alpha1.BuildSpec{
					SourcePackage: corev1.LocalObjectReference{
						Name: "some-package",
					},
				},
			},
		},
	}

	testCases := []struct {
		name       string
		sourcePath string
		setup      func(t *testing.T, f *fakes)
		assert     func(t *testing.T, err error)
	}{
		{
			name: "happy path",
			setup: func(t *testing.T, f *fakes) {
				f.cs.Fake.PrependReactor("create", "sourcepackages",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						sp := action.(ktesting.CreateAction).GetObject().(*v1alpha1.SourcePackage)

						testutil.AssertEqual(t, "name", app.Spec.Build.Spec.SourcePackage.Name, sp.Name)
						testutil.AssertEqual(t, "namespace", app.Namespace, sp.Namespace)
						testutil.AssertEqual(t, "len(OwnerReferences)", 1, len(sp.OwnerReferences))
						testutil.AssertEqual(t, "size", uint64(2560), sp.Spec.Size)
						testutil.AssertEqual(t, "checksum.type", v1alpha1.PackageChecksumSHA256Type, sp.Spec.Checksum.Type)
						testutil.AssertEqual(t, "checksum.value", "e6d4f054de9d1220a7e02da5e5fc2d4016c01404dda56d56d1dbeb18521dc4a3", sp.Spec.Checksum.Value)

						return true, sp, nil
					},
				)

				f.poster = func(
					ctx context.Context,
					requestURI string,
					bodyFileName string,
				) error {
					testutil.AssertEqual(t, "requestURI", "/apis/upload.kf.dev/v1alpha1/proxy/namespaces/some-namespace/some-package", requestURI)
					testutil.AssertNotBlank(t, "bodyFileName", bodyFileName)
					return nil
				}
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertNil(t, "err", err)
			},
		},
		{
			name:       "invalid source path",
			sourcePath: filepath.Join("testdata", "invalid"),
			assert: func(t *testing.T, err error) {
				// We can't directly assert on the error because the
				// underlying path depends on the machine.
				if err == nil || !strings.Contains(err.Error(), "no such file or directory") {
					t.Fatalf("expected error suggesting it can't find the directory")
				}
			},
		},
		{
			name: "failed to create SourcePackage",
			setup: func(t *testing.T, f *fakes) {
				f.cs.Fake.PrependReactor("create", "sourcepackages",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("some-error")
					},
				)
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to create SourcePackage: some-error"), err)
			},
		},
		{
			name: "failed to post",
			setup: func(t *testing.T, f *fakes) {
				f.cs.Fake.PrependReactor("create", "sourcepackages",
					func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						sp := action.(ktesting.CreateAction).GetObject().(*v1alpha1.SourcePackage)
						return true, sp, nil
					},
				)

				f.poster = func(
					ctx context.Context,
					requestURI string,
					bodyFileName string,
				) error {
					return errors.New("some-error")
				}
			},
			assert: func(t *testing.T, err error) {
				testutil.AssertErrorsEqual(t, errors.New("failed to upload source directory: some-error"), err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakes{
				cs: fake.NewSimpleClientset(),
			}
			if tc.setup != nil {
				tc.setup(t, f)
			}

			if tc.sourcePath == "" {
				tc.sourcePath = "testdata"
			}

			c := NewClient(f.cs.KfV1alpha1(), f.poster)

			err := c.UploadSourcePath(context.Background(), tc.sourcePath, app)
			if tc.assert != nil {
				tc.assert(t, err)
			}
		})
	}
}

func TestNewPoster(t *testing.T) {
	t.Parallel()

	// XXX: This function is tested only via endToEnd tests.
	// Unfortunately, this is a bit tough to unit test because the Kubernetes
	// client returns an actual type (as opposed to an interface) that will
	// actually need an API server to talk to. The FakeCorev1
	// (https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/core/v1/fake#FakeCoreV1.RESTClient)
	// doesn't even attempt to be helpful and instead returns a nil type.
}
