package baremetal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	common "github.com/maximilien/bosh-softlayer-cpi/common"
	bm "github.com/maximilien/bosh-softlayer-cpi/softlayer/baremetal"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("BaremetalCreator", func() {
	var (
		softLayerClient *fakeslclient.FakeSoftLayerClient
		logger          boshlog.Logger
		creator         bm.BaremetalCreator
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		logger = boshlog.NewLogger(boshlog.LevelNone)

		creator = bm.NewBaremetalCreator(
			softLayerClient,
			logger,
		)
		common.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Hardware_Service_createObject.json")
	})

	Describe("Create", func() {
		Context("succeeded", func() {
			It("returns a new Softlayer Hardware without an error", func() {
				baremetal, err := creator.Create(2, 1, 10, "fake-host", "fake-domain", "fake-os", "fake-data-center")
				Expect(err).ToNot(HaveOccurred())
				Expect(baremetal.GlobalIdentifier).To(Equal("fake-id"))
			})
		})
	})
})
