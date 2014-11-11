package disk_test

import (
	// "errors"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "github.com/maximilien/bosh-softlayer-cpi/softlayer/cli/fakes"
	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
)

var _ = FDescribe("SoftLayerCreator", func() {
	var (
		softLayerClient *fakeslclient.Fake_SoftLayer_Client
		logger          boshlog.Logger
		creator         SoftLayerCreator
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFake_SoftLayer_Client()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		creator = NewSoftLayerCreator(softLayerClient, logger)
	})

	Describe("Create", func() {
		Context("When parameters are valid", func() {
			FIt("creates and return the new disk", func() {
				disk, err := creator.Create(20, 123)
				Expect(err).ToNot(HaveOccurred())
				Expect(disk.ID()).To(Equal(123))
			})
			FIt("creates and return the new disk when we don't specific the location", func() {
				disk, err := creator.Create(20, 0)
				Expect(err).ToNot(HaveOccurred())
				Expect(disk.ID()).To(Equal(123))
			})
		})

		Context("When parameters are invalid", func() {
			FIt("returns error", func() {
				_, err := creator.Create(-10, 0)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
