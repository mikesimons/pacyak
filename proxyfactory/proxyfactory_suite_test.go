package proxyfactory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestProxyfactory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Proxyfactory Suite")
}
