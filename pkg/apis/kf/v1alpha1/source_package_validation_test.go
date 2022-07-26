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

package v1alpha1

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	logtesting "knative.dev/pkg/logging/testing"
)

func TestPackage_Validate(t *testing.T) {
	t.Parallel()

	// Go won't let you slice an unaddressable value, so we have to save it to
	// a variable first.
	cz := sha256.Sum256([]byte("hello world\n"))
	goodChecksum := hex.EncodeToString(cz[:])

	goodSpec := SourcePackageSpec{
		Size: maxUploadSizeBytes,
		Checksum: SourcePackageChecksum{
			Type:  PackageChecksumSHA256Type,
			Value: goodChecksum,
		},
	}

	badMeta := metav1.ObjectMeta{
		Name: strings.Repeat("A", 64), // Too long
	}

	cases := map[string]struct {
		sut  SourcePackage
		want *apis.FieldError
		ctx  context.Context
	}{
		"valid spec": {
			sut: SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: goodSpec,
			},
		},
		"skips validation in status update": {
			sut: SourcePackage{
				ObjectMeta: badMeta,
				Spec:       goodSpec,
			},
			ctx:  apis.WithinSubResourceUpdate(context.Background(), nil, "status"),
			want: nil,
		},
		"invalid ObjectMeta": {
			sut: SourcePackage{
				ObjectMeta: badMeta,
				Spec:       goodSpec,
			},
			want: apis.ValidateObjectMetadata(badMeta.GetObjectMeta()).ViaField("metadata"),
		},
		"size too small": {
			sut: SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourcePackageSpec{
					Size: 0,
					Checksum: SourcePackageChecksum{
						Type:  PackageChecksumSHA256Type,
						Value: goodChecksum,
					},
				},
			},
			want: apis.ErrOutOfBoundsValue(0, minUploadSizeBytes, maxUploadSizeBytes, "spec.size"),
		},
		"size is too large": {
			sut: SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourcePackageSpec{
					Size: maxUploadSizeBytes + 1,
					Checksum: SourcePackageChecksum{
						Type:  PackageChecksumSHA256Type,
						Value: goodChecksum,
					},
				},
			},
			want: apis.ErrOutOfBoundsValue(maxUploadSizeBytes+1, minUploadSizeBytes, maxUploadSizeBytes, "spec.size"),
		},
		"invalid checksum type": {
			sut: SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourcePackageSpec{
					Size: 99,
					Checksum: SourcePackageChecksum{
						Type:  "invalid",
						Value: goodChecksum,
					},
				},
			},
			want: ErrInvalidEnumValue("invalid", "spec.checksum.type", []string{PackageChecksumSHA256Type}),
		},
		"invalid checksum value, not hex": {
			sut: SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourcePackageSpec{
					Size: 99,
					Checksum: SourcePackageChecksum{
						Type:  PackageChecksumSHA256Type,
						Value: strings.Repeat("J", 64),
					},
				},
			},
			want: apis.ErrInvalidValue(strings.Repeat("J", 64), "spec.checksum.value"),
		},
		"invalid checksum value, too short": {
			sut: SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourcePackageSpec{
					Size: 99,
					Checksum: SourcePackageChecksum{
						Type:  PackageChecksumSHA256Type,
						Value: hex.EncodeToString(bytes.Repeat([]byte{'a'}, 31)),
					},
				},
			},
			want: apis.ErrInvalidValue(hex.EncodeToString(bytes.Repeat([]byte{'a'}, 31)), "spec.checksum.value"),
		},
		"invalid checksum value, too long": {
			sut: SourcePackage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid",
				},
				Spec: SourcePackageSpec{
					Size: 99,
					Checksum: SourcePackageChecksum{
						Type:  PackageChecksumSHA256Type,
						Value: hex.EncodeToString(bytes.Repeat([]byte{'a'}, 33)),
					},
				},
			},
			want: apis.ErrInvalidValue(hex.EncodeToString(bytes.Repeat([]byte{'a'}, 33)), "spec.checksum.value"),
		},
	}

	store := config.NewDefaultConfigStore(logtesting.TestLogger(t))

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			ctx := context.Background()
			if tc.ctx != nil {
				ctx = tc.ctx
			}

			ctx = store.ToContext(ctx)
			got := tc.sut.Validate(ctx)

			testutil.AssertEqual(t, "validation errors", tc.want.Error(), got.Error())
		})
	}
}
