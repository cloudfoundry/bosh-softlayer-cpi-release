package disk_test

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
	. "github.com/onsi/ginkgo"
)

var _ = XDescribe("SoftLayerDisk", func() {
	var (
		disk SoftLayerDisk
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		disk = NewSoftLayerDisk(1234, logger)
	})

	Describe("Delete", func() {
		It("deletes an iSCSI disk successfully", func() {
		})

		It("returns error if deleting fails", func() {
		})
	})
})
