package secrets

import "fmt"

// BrokerCredentialSecretName creates a deterministic secret name from a broker name.
func BrokerCredentialSecretName(brokerName string) string {
	return fmt.Sprintf("service-broker-%s-creds", brokerName)
}
