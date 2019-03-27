package fake

import "github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks"

//go:generate mockgen --package=fake --destination=fake_builder_creator.go --mock_names=BuilderCreator=FakeBuilderCreator github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks/fake BuilderCreator
//go:generate mockgen --package=fake --destination=fake_build_template_uploader.go --mock_names=BuildTemplateUploader=FakeBuildTemplateUploader github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks/fake BuildTemplateUploader
//go:generate mockgen --package=fake --destination=fake_buildpack_lister.go --mock_names=BuildpackLister=FakeBuildpackLister github.com/GoogleCloudPlatform/kf/pkg/kf/buildpacks/fake BuildpackLister

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
