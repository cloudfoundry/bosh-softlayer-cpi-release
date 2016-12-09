package disk_test

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"

	fakeclient "github.com/maximilien/softlayer-go/client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
)

var _ = Describe("SoftLayerFinder", func() {
	var (
		fc     *fakeclient.FakeSoftLayerClient
		logger boshlog.Logger
		finder SoftLayerFinder
	)

	BeforeEach(func() {
		fc = fakeclient.NewFakeSoftLayerClient("fake-user", "fake-key")
		logger = boshlog.NewLogger(boshlog.LevelNone)
		finder = NewSoftLayerDiskFinder(fc, logger)
	})

	Describe("Find", func() {
		It("returns disk and found as true when found the disk successfully", func() {
			testhelpers.SetTestFixtureForFakeSoftLayerClient(fc, "SoftLayer_Network_Storage_Service_getIscsiVolume.json")

			disk, found, err := finder.Find(1234)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			expectedDisk := NewSoftLayerDisk(1234, fc, logger)
			Expect(disk).To(Equal(expectedDisk))
		})

		It("returns found as false when failed to find the disk", func() {
			testhelpers.SetTestFixtureForFakeSoftLayerClient(fc, "SoftLayer_Network_Storage_Service_getEmptyIscsiVolume.json")
			disk, found, err := finder.Find(1234)

			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
			Expect(disk).To(BeNil())
		})
	})
})
