package create_vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCreateVm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration: create_vm Suite")
}
