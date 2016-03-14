package create_disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCreateDisk(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration: create_disk Suite")
}
