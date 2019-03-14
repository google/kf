package fake

import "github.com/GoogleCloudPlatform/kf/pkg/kf/services"

//go:generate mockgen --package=fake --destination=fake_client_interface.go --mock_names=ClientInterface=FakeClientInterface github.com/GoogleCloudPlatform/kf/pkg/kf/services/fake ClientInterface

// ClientInterface is implementd by services.Client.
type ClientInterface interface {
	services.ClientInterface
}
