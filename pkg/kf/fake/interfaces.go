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

//go:generate mockgen --package=fake --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_lister.go --mock_names=Lister=FakeLister github.com/GoogleCloudPlatform/kf/pkg/kf/fake Lister
//go:generate mockgen --package=fake --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_pusher.go --mock_names=Pusher=FakePusher github.com/GoogleCloudPlatform/kf/pkg/kf/fake Pusher
//go:generate mockgen --package=fake --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_deleter.go --mock_names=Deleter=FakeDeleter github.com/GoogleCloudPlatform/kf/pkg/kf/fake Deleter
//go:generate mockgen --package=fake --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_log_tailer.go --mock_names=LogTailer=FakeLogTailer github.com/GoogleCloudPlatform/kf/pkg/kf/fake LogTailer
//go:generate mockgen --package=fake --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_environment_client.go --mock_names=EnvironmentClient=FakeEnvironmentClient github.com/GoogleCloudPlatform/kf/pkg/kf/fake EnvironmentClient

import (
	"github.com/GoogleCloudPlatform/kf/pkg/kf"
)

// Lister is implemented by kf.Lister.
type Lister interface {
	kf.AppLister
}

// Lister is implemented by kf.Pusher.
type Pusher interface {
	Push(appName string, opts ...kf.PushOption) error
}

// Lister is implemented by kf.Deleter.
type Deleter interface {
	Delete(appName string, opts ...kf.DeleteOption) error
}

// LogTailer is implemented by kf.LogTailer.
type LogTailer interface {
	kf.Logs
}

// EnvironmentClient is implemented by kf.EnvironmentClient
type EnvironmentClient interface {
	// List shows all the names and values of the environment variables for an
	// app.
	List(appName string, opts ...kf.ListEnvOption) (map[string]string, error)

	// Set sets the given environment variables.
	Set(appName string, values map[string]string, opts ...kf.SetEnvOption) error

	// Unset unsets the given environment variables.
	Unset(appName string, names []string, opts ...kf.UnsetEnvOption) error
}
