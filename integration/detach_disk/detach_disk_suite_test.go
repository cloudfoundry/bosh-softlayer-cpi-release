package detach_disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDetachDisk(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration: detach_disk Suite")
}
