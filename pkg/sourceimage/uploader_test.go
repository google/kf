package sourceimage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-containerregistry/pkg/name"
	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen --package=sourceimage --copyright_file ../kf/internal/tools/option-builder/LICENSE_HEADER --destination=fake_listers.go --mock_names=SpaceLister=FakeSpaceLister,SourcePackageLister=FakeSourcePackageLister,SourcePackageNamespaceLister=FakeSourcePackageNamespaceLister github.com/google/kf/v2/pkg/client/kf/listers/kf/v1alpha1 SpaceLister,SourcePackageLister,SourcePackageNamespaceLister

func TestUploader(t *testing.T) {
	t.Parallel()

	type fakes struct {
		fs  *FakeSpaceLister
		fp  *FakeSourcePackageLister
		fnp *FakeSourcePackageNamespaceLister
	}

	type output struct {
		ctrl                 *gomock.Controller
		image                name.Reference
		err                  error
		imagePusherInvoked   bool
		statusUpdaterInvoked bool
	}

	var (
		normalData             = "normal-data"
		normalChecksumValueSum = sha256.Sum256([]byte(normalData))
		normalChecksumValue    = hex.EncodeToString(normalChecksumValueSum[:])
	)

	setup := func(f fakes, spaceName, sourcePackageName, checksumType string, size uint64) {
		f.fs.EXPECT().
			Get(spaceName).
			Return(&v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					BuildConfig: v1alpha1.SpaceStatusBuildConfig{
						ContainerRegistry: "some-registry",
					},
				},
			}, nil)

		f.fp.EXPECT().
			SourcePackages(spaceName).
			Return(f.fnp)

		f.fnp.EXPECT().
			Get(sourcePackageName).
			Return(&v1alpha1.SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					UID: "some-uid",
				},
				Spec: v1alpha1.SourcePackageSpec{
					Size: size,
					Checksum: v1alpha1.SourcePackageChecksum{
						Type:  checksumType,
						Value: normalChecksumValue,
					},
				},
			}, nil)
	}

	setupWithSize := func(f fakes, spaceName, sourcePackageName string, size uint64) {
		setup(f, spaceName, sourcePackageName, v1alpha1.PackageChecksumSHA256Type, size)
	}

	setupNormal := func(f fakes, spaceName, sourcePackageName string) {
		setupWithSize(f, spaceName, sourcePackageName, uint64(len(normalData)))
	}

	testCases := []struct {
		name              string
		spaceName         string
		sourcePackageName string
		maxRetriesForGetSourcePackage int
		data              io.Reader
		setup             func(t *testing.T, f fakes)
		assert            func(t *testing.T, o output)

		imagePusher   func(t *testing.T, path, imageName string) (name.Reference, error)
		statusUpdater func(t *testing.T, s *v1alpha1.SourcePackage) error
	}{
		{
			name:              "pushes the correct image",
			spaceName:         "some-space",
			sourcePackageName: "some-name",
			maxRetriesForGetSourcePackage: 0,
			setup: func(t *testing.T, f fakes) {
				setupNormal(f, "some-space", "some-name")
			},
			data: strings.NewReader(normalData),
			imagePusher: func(t *testing.T, path, imageName string) (name.Reference, error) {
				data, err := ioutil.ReadFile(path)
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "data", normalData, string(data))
				testutil.AssertEqual(t, "imageName", "some-registry/some-uid", imageName)
				return name.NewTag("some-registry/some-uid")
			},
			statusUpdater: func(t *testing.T, s *v1alpha1.SourcePackage) error {
				testutil.AssertEqual(t, "image", "index.docker.io/some-registry/some-uid:latest", s.Status.Image)
				testutil.AssertEqual(t, "checksum.type", v1alpha1.PackageChecksumSHA256Type, s.Status.Checksum.Type)
				testutil.AssertEqual(t, "checksum.value", normalChecksumValue, s.Status.Checksum.Value)
				testutil.AssertEqual(t, "size", uint64(len(normalData)), s.Status.Size)
				testutil.AssertEqual(t, "succeeded", true, s.Status.Succeeded())
				return nil
			},
			assert: func(t *testing.T, o output) {
				testutil.AssertNil(t, "err", o.err)
				testutil.AssertEqual(t, "imagePusherInvoked", true, o.imagePusherInvoked)
				testutil.AssertEqual(t, "statusUpdaterInvoked", true, o.statusUpdaterInvoked)

				d, err := name.NewTag("some-registry/some-uid")
				testutil.AssertNil(t, "err", err)
				testutil.AssertEqual(t, "image", d, o.image)
			},
		},
		{
			name:      "getting Space fails",
			spaceName: "some-space",
			maxRetriesForGetSourcePackage: 0,
			setup: func(t *testing.T, f fakes) {
				f.fs.EXPECT().
					Get("some-space").
					Return(nil, errors.New("some-error"))
			},
			assert: func(t *testing.T, o output) {
				testutil.AssertErrorsEqual(t, errors.New("failed to find Space: some-error"), o.err)
			},
		},
		{
			name:              "getting SourcePackage fails",
			spaceName:         "some-space",
			sourcePackageName: "some-name",
			maxRetriesForGetSourcePackage: 4,
			setup: func(t *testing.T, f fakes) {
				f.fs.EXPECT().Get("some-space")

				f.fp.EXPECT().
					SourcePackages("some-space").
					Return(f.fnp).Times(5) // One initial attempt and then 4 retries.

				f.fnp.EXPECT().
					Get("some-name").
					Return(nil, errors.New("some-error")).Times(5) // One initial attempt and then 4 retries.
			},
			assert: func(t *testing.T, o output) {
				testutil.AssertErrorsEqual(t, errors.New("failed to find SourcePackage, retries exhausted: some-error"), o.err)
			},
		},
		{
			name:              "SourcePackage is not pending",
			spaceName:         "some-space",
			sourcePackageName: "some-name",
			maxRetriesForGetSourcePackage: 0,
			setup: func(t *testing.T, f fakes) {
				f.fs.EXPECT().Get("some-space")

				f.fp.EXPECT().
					SourcePackages("some-space").
					Return(f.fnp)

				status := v1alpha1.SourcePackageStatus{}
				status.PropagateSpec("some-image", v1alpha1.SourcePackageSpec{})

				f.fnp.EXPECT().
					Get("some-name").
					Return(&v1alpha1.SourcePackage{
						ObjectMeta: metav1.ObjectMeta{
							UID: "some-uid",
						},
						Spec: v1alpha1.SourcePackageSpec{
							Size: 99,
							Checksum: v1alpha1.SourcePackageChecksum{
								Type:  v1alpha1.PackageChecksumSHA256Type,
								Value: normalChecksumValue,
							},
						},
						Status: status,
					}, nil)
			},
			assert: func(t *testing.T, o output) {
				testutil.AssertErrorsEqual(t, errors.New("SourcePackage is not pending"), o.err)
			},
		},
		{
			name:              "building and pushing image fails",
			spaceName:         "some-space",
			sourcePackageName: "some-name",
			maxRetriesForGetSourcePackage: 0,
			data:              strings.NewReader(normalData),
			setup: func(t *testing.T, f fakes) {
				setupNormal(f, "some-space", "some-name")
			},
			imagePusher: func(t *testing.T, path, imageName string) (name.Reference, error) {
				return nil, errors.New("some-error")
			},
			assert: func(t *testing.T, o output) {
				testutil.AssertErrorsEqual(t, errors.New("failed to build and push image: some-error"), o.err)
			},
		},
		{
			name:              "updating status fails",
			spaceName:         "some-space",
			sourcePackageName: "some-name",
			maxRetriesForGetSourcePackage: 0,
			data:              strings.NewReader(normalData),
			setup: func(t *testing.T, f fakes) {
				setupNormal(f, "some-space", "some-name")
			},
			statusUpdater: func(t *testing.T, s *v1alpha1.SourcePackage) error {
				return errors.New("some-error")
			},
			assert: func(t *testing.T, o output) {
				testutil.AssertErrorsEqual(t, errors.New("failed to update SourcePackage status: some-error"), o.err)
			},
		},
		{
			name:              "saving data fails",
			spaceName:         "some-space",
			sourcePackageName: "some-name",
			maxRetriesForGetSourcePackage: 0,
			data:              &errReader{err: errors.New("some-error")},
			setup: func(t *testing.T, f fakes) {
				setupNormal(f, "some-space", "some-name")
			},
			assert: func(t *testing.T, o output) {
				testutil.AssertErrorsEqual(t, errors.New("failed to save data: some-error"), o.err)
			},
		},
		{
			name:              "checksum doesn't match",
			spaceName:         "some-space",
			sourcePackageName: "some-name",
			maxRetriesForGetSourcePackage: 0,

			// NOTE: the variable normalData is associated with
			// normalChecksumValue which is set by the setupNormal function.
			data: strings.NewReader("NORMAL-data"),
			setup: func(t *testing.T, f fakes) {
				setupNormal(f, "some-space", "some-name")
			},
			assert: func(t *testing.T, o output) {
				testutil.AssertErrorsEqual(t, errors.New("checksum does not match expected"), o.err)
			},
		},
		{
			name:              "size doesn't match spec",
			spaceName:         "some-space",
			sourcePackageName: "some-name",
			maxRetriesForGetSourcePackage: 0,
			data:              strings.NewReader(normalData),
			setup: func(t *testing.T, f fakes) {
				setupWithSize(f, "some-space", "some-name", uint64(len(normalData)+1))
			},
			assert: func(t *testing.T, o output) {
				testutil.AssertErrorsEqual(t, errors.New("expected 12 bytes, got 11"), o.err)
			},
		},
		{
			name:              "unknown checksum type",
			spaceName:         "some-space",
			sourcePackageName: "some-name",
			maxRetriesForGetSourcePackage: 0,
			data:              strings.NewReader(normalData),
			setup: func(t *testing.T, f fakes) {
				setup(f, "some-space", "some-name", "invalid", uint64(len(normalData)))
			},
			assert: func(t *testing.T, o output) {
				testutil.AssertErrorsEqual(t, errors.New("unknown checksum type: invalid"), o.err)
			},
		},
		{
			name:              "limits how much it reads",
			spaceName:         "some-space",
			sourcePackageName: "some-name",
			maxRetriesForGetSourcePackage: 0,
			// This reader will never stop feeding the test data. Unless it
			// has properly guarded itself by limiting how much data it reads,
			// it will timeout.
			data: &neverEndingReader{},
			setup: func(t *testing.T, f fakes) {
				setupWithSize(f, "some-space", "some-name", uint64(len(normalData)+1))
			},
			assert: func(t *testing.T, o output) {
				testutil.AssertErrorsEqual(t, errors.New("checksum does not match expected"), o.err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fs := NewFakeSpaceLister(ctrl)
			fp := NewFakeSourcePackageLister(ctrl)
			fnp := NewFakeSourcePackageNamespaceLister(ctrl)

			if tc.imagePusher == nil {
				tc.imagePusher = func(t *testing.T, path, imageName string) (name.Reference, error) {
					// NOP
					return name.NewTag(imageName)
				}
			}

			if tc.statusUpdater == nil {
				tc.statusUpdater = func(t *testing.T, s *v1alpha1.SourcePackage) error {
					// NOP
					return nil
				}
			}

			var imagePusherInvoked, statusUpdaterInvoked bool

			u := NewUploader(fs, fp,
				// Create an image pusher function.
				func(path, imageName string) (name.Reference, error) {
					imagePusherInvoked = true
					return tc.imagePusher(t, path, imageName)
				},
				// Create a status updater function.
				func(s *v1alpha1.SourcePackage) error {
					statusUpdaterInvoked = true
					return tc.statusUpdater(t, s)
				},
			)

			if tc.setup != nil {
				tc.setup(t, fakes{
					fs:  fs,
					fp:  fp,
					fnp: fnp,
				})
			}

			image, err := u.Upload(
				context.Background(),
				tc.spaceName,
				tc.sourcePackageName,
				tc.maxRetriesForGetSourcePackage,
				tc.data,
			)

			if tc.assert != nil {
				tc.assert(t, output{
					ctrl:                 ctrl,
					image:                image,
					err:                  err,
					imagePusherInvoked:   imagePusherInvoked,
					statusUpdaterInvoked: statusUpdaterInvoked,
				})
			}
		})
	}
}

type errReader struct {
	err error
}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

type neverEndingReader struct{}

func (n *neverEndingReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'x'
	}
	return len(p), nil
}
