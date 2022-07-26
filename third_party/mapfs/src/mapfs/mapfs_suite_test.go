package mapfs_test

import (
. "github.com/onsi/ginkgo"
. "github.com/onsi/gomega"

"testing"
)

func TestBroker(t *testing.T) {
RegisterFailHandler(Fail)
RunSpecs(t, "MapFS Suite")
}
