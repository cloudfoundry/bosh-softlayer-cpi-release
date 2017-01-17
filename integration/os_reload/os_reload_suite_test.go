package os_reload_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCreateVm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration: os_reload Suite")
}
