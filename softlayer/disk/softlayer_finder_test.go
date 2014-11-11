package disk_test

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	// fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
)

var _ = Describe("SoftLayerFinder", func() {
	var (
		logger boshlog.Logger
		finder SoftLayerFinder
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		finder = NewSoftLayerFinder(logger)
	})

	FDescribe("Find", func() {
		It("returns disk and found as true if disk path exists", func() {
		})

		It("returns found as false if disk path does not exist", func() {
		})
	})
})
