package attach_disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAttachDisk(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration: attach_disk Suite")
}
