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
//go:generate mockgen --package=fake --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_log_tailer.go --mock_names=LogTailer=FakeLogTailer github.com/GoogleCloudPlatform/kf/pkg/kf/fake LogTailer
//go:generate mockgen --package=fake --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_istio_client.go --mock_names=IstioClient=FakeIstioClient github.com/GoogleCloudPlatform/kf/pkg/kf/fake IstioClient
//go:generate mockgen --package=fake --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_systemenvinjector.go --mock_names=SystemEnvInjector=FakeSystemEnvInjector github.com/GoogleCloudPlatform/kf/pkg/kf/fake SystemEnvInjector
//go:generate mockgen --package=fake --copyright_file ../internal/tools/option-builder/LICENSE_HEADER --destination=fake_deployer.go --mock_names=Deployer=FakeDeployer github.com/GoogleCloudPlatform/kf/pkg/kf/fake Deployer

import (
	"github.com/GoogleCloudPlatform/kf/pkg/kf"
)

// Lister is implemented by kf.Lister.
type Lister interface {
	kf.AppLister
}

// Pusher is implemented by kf.Pusher.
type Pusher interface {
	kf.Pusher
}

// LogTailer is implemented by kf.LogTailer.
type LogTailer interface {
	kf.Logs
}

// IstioClient is implemented by kf.IstioClient.
type IstioClient interface {
	kf.IngressLister
}

// SystemEnvInjector is implemented by kf.SystemEnvInjector
type SystemEnvInjector interface {
	kf.SystemEnvInjectorInterface
}

// Deployer is implemented by kf.Deployer
type Deployer interface {
	kf.Deployer
}
