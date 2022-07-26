package fake

import "github.com/google/kf/v2/pkg/kf/sourcepackages"

//go:generate mockgen --package=fake --copyright_file ../../internal/tools/option-builder/LICENSE_HEADER --destination=fake_client.go --mock_names=Client=FakeClient github.com/google/kf/v2/pkg/kf/sourcepackages/fake Client

// Client is the client for spaces.
type Client interface {
	sourcepackages.Client
}
