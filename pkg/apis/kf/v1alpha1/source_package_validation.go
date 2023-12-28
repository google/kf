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
	"context"
	"encoding/hex"

	"knative.dev/pkg/apis"
)

const (
	minUploadSizeBytes = 1
	maxUploadSizeBytes = 3 * 1024 * 1024 * 1024
)

// Validate checks for errors in the SourcePackage's spec or status fields.
func (p *SourcePackage) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	errs = errs.Also(apis.ValidateObjectMetadata(p.GetObjectMeta()).ViaField("metadata"))
	errs = errs.Also(p.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))

	return errs
}

// Validate makes sure that a SourcePackageSpec is properly configured.
func (spec *SourcePackageSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	if spec.Size < minUploadSizeBytes || spec.Size > maxUploadSizeBytes {
		errs = errs.Also(apis.ErrOutOfBoundsValue(
			spec.Size, minUploadSizeBytes, maxUploadSizeBytes, "size"),
		)
	}
	errs = errs.Also(spec.Checksum.Validate(apis.WithinSpec(ctx)).ViaField("checksum"))

	return errs
}

// Validate makes sure that a SourcePackageChecksum is properly configured.
func (c *SourcePackageChecksum) Validate(ctx context.Context) (errs *apis.FieldError) {
	// Only sha256 is allowed at the moment. This *might* change in the
	// future.
	if c.Type != PackageChecksumSHA256Type {
		errs = errs.Also(
			ErrInvalidEnumValue(
				"invalid",
				"type",
				[]string{PackageChecksumSHA256Type},
			),
		)
	}

	hexData, err := hex.DecodeString(c.Value)
	if err != nil || len(hexData) != 32 {
		errs = apis.ErrInvalidValue(c.Value, "value")
	}
	return errs
}
