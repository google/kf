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
	"testing"

	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestSourcePackageStatus_PropagateSpec(t *testing.T) {
	t.Parallel()

	status := &SourcePackageStatus{}
	status.PropagateSpec("some-image", SourcePackageSpec{
		Checksum: SourcePackageChecksum{
			Type:  "some-checksum",
			Value: "some-value",
		},
		Size: 99,
	})

	testutil.AssertEqual(t, "image", "some-image", status.Image)
	testutil.AssertEqual(t, "checksum.type", "some-checksum", status.Checksum.Type)
	testutil.AssertEqual(t, "checksum.value", "some-value", status.Checksum.Value)
	testutil.AssertEqual(t, "size", uint64(99), status.Size)
	testutil.AssertEqual(t, "succeeded", true, status.Succeeded())
}
