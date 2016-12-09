package disk_test

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"

	fakeclient "github.com/maximilien/softlayer-go/client/fakes"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SoftLayerDisk", func() {
	var (
		fc   *fakeclient.FakeSoftLayerClient
		disk SoftLayerDisk
	)

	BeforeEach(func() {
		fc = fakeclient.NewFakeSoftLayerClient("fake-user", "fake-key")
		logger := boshlog.NewLogger(boshlog.LevelNone)
		disk = NewSoftLayerDisk(1234, fc, logger)
	})

	Describe("Delete", func() {
		It("deletes an iSCSI disk successfully", func() {
			fileNames := []string{
				"SoftLayer_Network_Storage_Service_getBillingItem.json",
				"SoftLayer_Billing_Item_Service_cancelService.json",
			}
			testhelpers.SetTestFixturesForFakeSoftLayerClient(fc, fileNames)

			err := disk.Delete()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
