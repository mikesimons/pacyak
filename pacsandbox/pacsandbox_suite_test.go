package pacsandbox_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPacsandbox(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pacsandbox Suite")
}
