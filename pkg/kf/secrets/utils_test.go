package secrets

import "fmt"

func ExampleBrokerCredentialSecretName() {
	fmt.Println(BrokerCredentialSecretName("my-broker"))

	// Output: service-broker-my-broker-creds
}
