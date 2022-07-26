package fake

import "github.com/google/kf/v2/pkg/kf/tasks"

//go:generate mockgen --package=fake --copyright_file ../../internal/tools/option-builder/LICENSE_HEADER --destination=fake_client.go --mock_names=Client=FakeClient github.com/google/kf/v2/pkg/kf/tasks/fake Client

// Client is the client for Tasks.
type Client interface {
	tasks.Client
}
