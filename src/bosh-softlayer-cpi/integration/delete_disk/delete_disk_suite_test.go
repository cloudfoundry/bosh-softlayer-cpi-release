package delete_disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDeleteDisk(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration: delete_disk Suite")
}
