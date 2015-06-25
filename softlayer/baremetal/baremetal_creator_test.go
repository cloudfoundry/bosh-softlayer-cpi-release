package baremetal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bm "github.com/maximilien/bosh-softlayer-cpi/softlayer/baremetal"
	testhelpers "github.com/maximilien/bosh-softlayer-cpi/test_helpers"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

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
		testhelpers.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Hardware_Service_createObject.json")
	})

	Describe("Create", func() {
		Context("succeeded", func() {
			It("returns a new Softlayer Hardware without an error", func() {
				baremetal, err := creator.Create(2, 1, 10, "fake-host", "fake-domain", "fake-os", "fake-data-center")
				Expect(err).ToNot(HaveOccurred())
				Expect(baremetal.GlobalIdentifier).To(Equal("fake-id"))
				Expect(baremetal.BareMetalInstanceFlag).To(Equal(0))
				Expect(baremetal.ProvisionDate).To(BeNil())
				Expect(baremetal.PrimaryIpAddress).To(Equal(""))
			})
		})

		Context("failed", func() {
			Context("due to incorrect arguments", func() {
				It("returns an error when memory is incorrect", func() {
					_, err := creator.Create(-2, 1, 10, "fake-host", "fake-domain", "fake-os", "fake-data-center")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("memory can not be negative: -2"))
				})

				It("returns an error when processor is incorrect", func() {
					_, err := creator.Create(2, -1, 10, "fake-host", "fake-domain", "fake-os", "fake-data-center")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("processor can not be negative: -1"))
				})

				It("returns an error when disksize is incorrect", func() {
					_, err := creator.Create(2, 1, -10, "fake-host", "fake-domain", "fake-os", "fake-data-center")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("disk size can not be negative: -10"))
				})

				It("returns an error when host is incorrect", func() {
					_, err := creator.Create(2, 1, 10, "", "fake-domain", "fake-os", "fake-data-center")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("host can not be empty."))
				})

				It("returns an error when domain is incorrect", func() {
					_, err := creator.Create(2, 1, 10, "fake-host", "", "fake-os", "fake-data-center")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("domain can not be empty."))
				})

				It("returns an error when ostype is incorrect", func() {
					_, err := creator.Create(2, 1, 10, "fake-host", "fake-domain", "", "fake-data-center")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("os type can not be empty."))
				})

				It("returns an error when datacenter is incorrect", func() {
					_, err := creator.Create(2, 1, 10, "fake-host", "fake-domain", "fake-os", "")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("data center can not be empty."))
				})
			})
		})
	})
})
