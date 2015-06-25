package disk_test

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "github.com/maximilien/bosh-softlayer-cpi/test_helpers"

	fakeclient "github.com/maximilien/softlayer-go/client/fakes"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
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
				"SoftLayer_Account_Service_getIscsiVolume.json",
				"SoftLayer_Billing_Item_Cancellation_Request_Service_createObject.json",
			}
			testhelpers.SetTestFixturesForFakeSoftLayerClient(fc, fileNames)

			err := disk.Delete()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
