package buildpacks

import (
	"archive/tar"
	"io"

	"github.com/buildpack/lifecycle"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	gcrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	toml "github.com/pelletier/go-toml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildpackLister lists the available buildpacks. It should be created via
// NewBuildpackLister.
type BuildpackLister struct {
	f   BuildFactory
	rif RemoteImageFetcher
}

// RemoteImageFetcher is implemented by
// github.com/google/go-containerregistry/pkg/v1/remote.Image
type RemoteImageFetcher func(ref name.Reference, options ...remote.ImageOption) (gcrv1.Image, error)

// NewBuildpackLister creates a new BuildpackLister.
func NewBuildpackLister(f BuildFactory, rif RemoteImageFetcher) *BuildpackLister {
	return &BuildpackLister{
		f:   f,
		rif: rif,
	}
}

// List lists the available buildpacks.
func (l *BuildpackLister) List(opts ...BuildpackListOption) ([]string, error) {
	cfg := BuildpackListOptionDefaults().Extend(opts).toConfig()
	c, err := l.f()
	if err != nil {
		return nil, err
	}
	templates, err := c.BuildTemplates(cfg.Namespace).List(metav1.ListOptions{
		FieldSelector: "metadata.name=buildpack",
	})
	if err != nil {
		return nil, err
	}

	if len(templates.Items) == 0 {
		return nil, nil
	}

	builderImage := l.fetchBuilderImageName(templates.Items[0].Spec.Parameters)

	imageRef, err := name.ParseReference(builderImage, name.WeakValidation)
	if err != nil {
		return nil, err
	}

	image, err := l.rif(imageRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, err
	}

	ls, err := image.Layers()
	if err != nil {
		return nil, err
	}

	for i := len(ls) - 1; i >= 0; i-- {
		layer := ls[i]
		tr, closer, err := l.fetchImageTar(layer)
		if err != nil {
			return nil, err
		}
		defer closer.Close()

		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			if header.Name == "/buildpacks/order.toml" {
				return l.readOrder(tr)
			}
		}
	}

	return nil, nil
}

func (l *BuildpackLister) readOrder(reader io.Reader) ([]string, error) {
	var buildpackIDs []string
	var order struct {
		Groups []lifecycle.BuildpackGroup `toml:"groups"`
	}
	if err := toml.NewDecoder(reader).Decode(&order); err != nil {
		return nil, err
	}

	for _, group := range order.Groups {
		for _, bp := range group.Buildpacks {
			buildpackIDs = append(buildpackIDs, bp.ID)
		}
	}

	return buildpackIDs, nil
}

func (l *BuildpackLister) fetchImageTar(layer gcrv1.Layer) (*tar.Reader, io.Closer, error) {
	ucl, err := layer.Uncompressed()
	if err != nil {
		return nil, nil, err
	}

	return tar.NewReader(ucl), ucl, nil
}

func (l *BuildpackLister) fetchBuilderImageName(params []build.ParameterSpec) string {
	for _, p := range params {
		if p.Name == "BUILDER_IMAGE" && p.Default != nil {
			return *p.Default
		}
	}

	return ""
}
