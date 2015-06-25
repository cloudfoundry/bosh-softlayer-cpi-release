package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	testhelpers "github.com/maximilien/bosh-softlayer-cpi/test_helpers"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	fakesslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("SoftLayerFinder", func() {
	var (
		softLayerClient *fakesslclient.FakeSoftLayerClient
		logger          boshlog.Logger
		finder          SoftLayerFinder
	)

	BeforeEach(func() {
		softLayerClient = fakesslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		testhelpers.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Account_Service_getVirtualDiskImages.json")

		logger = boshlog.NewLogger(boshlog.LevelNone)
		finder = NewSoftLayerFinder(softLayerClient, logger)
	})

	Describe("Find", func() {
		Context("valid stemcell ID pointing to a SL virtual disk image", func() {
			It("returns stemcell and found as true if stemcell ", func() {
				stemcell, found, err := finder.Find("8c7a8358-d9a9-4e4d-9345-6f637e10ccb7")
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())

				expectedStemcell := NewSoftLayerStemcell(4868344, "8c7a8358-d9a9-4e4d-9345-6f637e10ccb7", VirtualDiskImageKind, softLayerClient, logger)
				Expect(stemcell).To(Equal(expectedStemcell))
			})
		})

		Context("valid stemcell ID pointing to a SL virtual disk image", func() {
			It("returns found as false if stemcell directory does not exist", func() {
				stemcell, found, err := finder.Find("fake-stemcell-id")
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
				Expect(stemcell).To(BeNil())
			})
		})
	})
})
