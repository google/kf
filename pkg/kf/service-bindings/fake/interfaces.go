package fake

import servicebindings "github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings"

//go:generate mockgen --package=fake --destination=fake_client_interface.go --mock_names=ClientInterface=FakeClientInterface github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings/fake ClientInterface

// ClientInterface is implementd by servicebinidngs.Client.
type ClientInterface interface {
	servicebindings.ClientInterface
}
