package baremetal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBaremetal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Baremetal Suite")
}
