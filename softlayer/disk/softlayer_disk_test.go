package disk_test

import (
	// "errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	// fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = FDescribe("IscsiDisk", func() {
	var (
		disk IscsiDisk
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		disk = NewIscsiDisk(1234, logger)
	})

	Describe("Delete", func() {
		It("deletes iSCSI volume", func() {
		})

		It("returns error if deleting iSCSI volume fails", func() {
		})
	})
})
