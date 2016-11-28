package hardware_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHardware(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hardware Suite")
}
