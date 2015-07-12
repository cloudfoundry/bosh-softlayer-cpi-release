package disk_test

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "github.com/maximilien/bosh-softlayer-cpi/test_helpers"

	fakeclient "github.com/maximilien/softlayer-go/client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
)

var _ = Describe("SoftLayerCreator", func() {
	var (
		fc      *fakeclient.FakeSoftLayerClient
		logger  boshlog.Logger
		creator SoftLayerCreator
	)

	BeforeEach(func() {
		fc = fakeclient.NewFakeSoftLayerClient("fake-user", "fake-key")
		logger = boshlog.NewLogger(boshlog.LevelNone)
		creator = NewSoftLayerDiskCreator(fc, logger)
	})

	Describe("Create", func() {
		var (
			cloudProps DiskCloudProperties
		)

		Context("Creates disk successfully", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_getObject.json",
					"SoftLayer_Product_Order_Service_getItemPrices.json",
					"SoftLayer_Product_Order_Service_placeOrder.json",
					"SoftLayer_Account_Service_getIscsiVolume.json",
				}
				cloudProps = DiskCloudProperties{
					ConsistentPerformanceIscsi: true,
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClient(fc, fileNames)
			})

			It("creates disk successfully and returns unique disk id", func() {
				disk, err := creator.Create(20, cloudProps, 123)
				Expect(err).ToNot(HaveOccurred())

				expectedDisk := NewSoftLayerDisk(1234, fc, logger)
				Expect(disk).To(Equal(expectedDisk))
			})
		})

		Context("Failed to create disk", func() {
			It("Reports error due to wrong virtual guest id", func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_getEmptyObject.json",
					"SoftLayer_Product_Order_Service_getItemPrices.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClient(fc, fileNames)

				_, err := creator.Create(20, cloudProps, 0)
				Expect(err).To(HaveOccurred())
			})
		})

	})
})
