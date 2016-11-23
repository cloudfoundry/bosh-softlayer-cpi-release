package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"

	slhelper "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	fakesslclient "github.com/maximilien/softlayer-go/client/fakes"
	"time"
)

var _ = Describe("SoftLayerFinder", func() {
	var (
		softLayerClient  *fakesslclient.FakeSoftLayerClient
		logger           boshlog.Logger
		finder           SoftLayerStemcellFinder
		expectedStemcell SoftLayerStemcell
	)

	BeforeEach(func() {
		softLayerClient = fakesslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")

		slhelper.TIMEOUT = 10 * time.Millisecond
		slhelper.POLLING_INTERVAL = 2 * time.Millisecond
		logger = boshlog.NewLogger(boshlog.LevelNone)

		expectedStemcell = NewSoftLayerStemcell(200150, "8071601b-5ee1-483e-a9e8-6e5582dcb9f7", softLayerClient, logger)
	})

	Describe("FindById", func() {
		Context("Success if http code 200 returns from SL", func() {
			It("returns stemcell if stemcell exists", func() {
				testhelpers.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service_getObject.json")

				softLayerClient.FakeHttpClient.DoRawHttpRequestInt = 200
				finder = NewSoftLayerStemcellFinder(softLayerClient, logger)

				stemcell, err := finder.FindById(200150)
				Expect(err).ToNot(HaveOccurred())
				Expect(stemcell).To(Equal(expectedStemcell))
			})
		})

		Context("Failed if the stemcell does not exists, 404 error returned", func() {
			It("returns error if stemcell does not exist", func() {
				softLayerClient.FakeHttpClient.DoRawHttpRequestInt = 404
				finder = NewSoftLayerStemcellFinder(softLayerClient, logger)

				_, err := finder.FindById(200150)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
