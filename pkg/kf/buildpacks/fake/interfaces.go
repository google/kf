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

import "github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"

//go:generate mockgen --package=fake --copyright_file ../../internal/tools/option-builder/LICENSE_HEADER --destination=fake_builder_creator.go --mock_names=BuilderCreator=FakeBuilderCreator github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks/fake BuilderCreator
//go:generate mockgen --package=fake --copyright_file ../../internal/tools/option-builder/LICENSE_HEADER --destination=fake_build_template_uploader.go --mock_names=BuildTemplateUploader=FakeBuildTemplateUploader github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks/fake BuildTemplateUploader
//go:generate mockgen --package=fake --copyright_file ../../internal/tools/option-builder/LICENSE_HEADER --destination=fake_buildpack_lister.go --mock_names=BuildpackLister=FakeBuildpackLister github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks/fake BuildpackLister

// BuilderFactory is implemented by buildpacks.BuilderFactory.
type BuilderCreator interface {
	Create(dir, containerRegistry string) (string, error)
}

// BuildTemplateUploader is implemented by buildpacks.BuildTemplateUploader.
type BuildTemplateUploader interface {
	UploadBuildTemplate(imageName string, opts ...buildpacks.UploadBuildTemplateOption) error
}

// BuildpackLister is implemented by buildpacks.BuildpackLister.
type BuildpackLister interface {
	List(opts ...buildpacks.BuildpackListOption) ([]string, error)
}
