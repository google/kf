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

import "github.com/google/kf/pkg/kf/apps"

//go:generate mockgen --package=fake --copyright_file ../../internal/tools/option-builder/LICENSE_HEADER --destination=fake_client.go --mock_names=Client=FakeClient github.com/google/kf/pkg/kf/apps/fake Client
//go:generate mockgen --package=fake --copyright_file ../../internal/tools/option-builder/LICENSE_HEADER --destination=fake_pusher.go --mock_names=Pusher=FakePusher github.com/google/kf/pkg/kf/apps/fake Pusher

// Client is the client for spaces.
type Client interface {
	apps.Client
}

// Pusher is implemented by kf.Pusher.
type Pusher interface {
	apps.Pusher
}
