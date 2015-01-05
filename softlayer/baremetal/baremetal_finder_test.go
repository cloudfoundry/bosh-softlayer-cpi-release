package baremetal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	common "github.com/maximilien/bosh-softlayer-cpi/common"
	bm "github.com/maximilien/bosh-softlayer-cpi/softlayer/baremetal"
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
		common.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Hardware_Service_getObject.json")
	})

	Describe("Find", func() {
		Context("succeeded", func() {
			It("returns a new Softlayer Hardware without an error", func() {
				baremetal, err := finder.Find("fake-id")
				Expect(err).ToNot(HaveOccurred())
				Expect(baremetal.GlobalIdentifier).To(Equal("fake-id"))
			})
		})
	})
})
