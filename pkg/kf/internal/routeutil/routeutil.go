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

package routeutil

import (
	"encoding/base64"
	"hash/crc64"
	"strings"
)

// EncodeRouteName is used to create a valid DNS name that can be used as a
// name for something like a VirtualService.
func EncodeRouteName(hostname, domain, urlPath string) string {
	hasher := crc64.New(crc64.MakeTable(crc64.ECMA))
	return hostname + "-" +
		strings.ToLower(
			base64.RawURLEncoding.EncodeToString(
				hasher.Sum([]byte(hostname+domain+urlPath)),
			),
		)
}
