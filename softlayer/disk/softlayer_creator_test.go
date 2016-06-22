package disk_test

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"

	fakeclient "github.com/maximilien/softlayer-go/client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
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

		Context("Creates disk successfully with cloud properties", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Product_Order_Service_getItems.json",
					"SoftLayer_Product_Order_Service_getItemPrices.json",
					"SoftLayer_Product_Order_Service_getItemPricesBySizeAndIops.json",
					"SoftLayer_Product_Order_Service_placeOrder.json",
					"SoftLayer_Account_Service_getIscsiVolume.json",
				}
				cloudProps = DiskCloudProperties{
					Iops:             1000,
					UseHourlyPricing: true,
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

		Context("Creates disk successfully without cloud properties", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Product_Order_Service_getItems.json",
					"SoftLayer_Product_Order_Service_getItemPrices.json",
					"SoftLayer_Product_Order_Service_getIopsItemPrices.json",
					"SoftLayer_Product_Order_Service_placeOrder.json",
					"SoftLayer_Account_Service_getIscsiVolume.json",
				}
				cloudProps = DiskCloudProperties{}
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
