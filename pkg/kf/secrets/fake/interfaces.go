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

package fake

import "github.com/google/kf/pkg/kf/secrets"

//go:generate mockgen --package=fake --copyright_file ../../internal/tools/option-builder/LICENSE_HEADER --destination=fake_client_interface.go --mock_names=ClientInterface=FakeClientInterface github.com/google/kf/pkg/kf/secrets/fake ClientInterface

// ClientInterface is implementd by secrets.Client.
type ClientInterface interface {
	secrets.ClientInterface
}
