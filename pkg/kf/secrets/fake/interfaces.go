package fake

import "github.com/GoogleCloudPlatform/kf/pkg/kf/secrets"

//go:generate mockgen --package=fake --destination=fake_client_interface.go --mock_names=ClientInterface=FakeClientInterface github.com/GoogleCloudPlatform/kf/pkg/kf/secrets/fake ClientInterface

// ClientInterface is implementd by secrets.Client.
type ClientInterface interface {
	secrets.ClientInterface
}
