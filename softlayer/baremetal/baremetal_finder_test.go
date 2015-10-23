package baremetal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bm "github.com/maximilien/bosh-softlayer-cpi/softlayer/baremetal"
	testhelpers "github.com/maximilien/bosh-softlayer-cpi/test_helpers"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("BaremetalFinder", func() {
	var (
		softLayerClient *fakeslclient.FakeSoftLayerClient
		logger          boshlog.Logger
		finder          bm.BaremetalFinder
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		logger = boshlog.NewLogger(boshlog.LevelNone)

		finder = bm.NewBaremetalFinder(
			softLayerClient,
			logger,
		)
	})

	Describe("Find", func() {
		Context("succeeded", func() {
			BeforeEach(func() {
				testhelpers.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Hardware_Service_getObject.json")
			})

			It("returns a new Softlayer Hardware without an error", func() {
				baremetal, err := finder.Find("fake-id")
				Expect(err).ToNot(HaveOccurred())
				Expect(baremetal.GlobalIdentifier).To(Equal("fake-id"))
				Expect(baremetal.BareMetalInstanceFlag).To(Equal(1))
				Expect(baremetal.ProvisionDate).ToNot(BeNil())
				Expect(baremetal.PrimaryIpAddress).To(Equal("1.1.1.1"))
			})
		})

		Context("failed", func() {
			BeforeEach(func() {
				testhelpers.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Hardware_Service_getObject_None_Exist.json")
			})

			It("return an error when the specified hardward id can not be found", func() {
				_, err := finder.Find("none-exist-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("cannot find the baremetal server with id: none-exist-id."))
			})

		})
	})
})
