package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"

	slhelper "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"
	fakesslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("SoftLayerStemcell", func() {
	var (
		fakeSoftLayerClient *fakesslclient.FakeSoftLayerClient
		stemcell            SoftLayerStemcell
		logger              boshlog.Logger
	)

	BeforeEach(func() {
		fakeSoftLayerClient = fakesslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")

		logger = boshlog.NewLogger(boshlog.LevelNone)

		stemcell = NewSoftLayerStemcell(1234, "fake-stemcell-uuid", fakeSoftLayerClient, logger)

		slhelper.TIMEOUT = 10 * time.Millisecond
		slhelper.POLLING_INTERVAL = 2 * time.Millisecond
	})

	Describe("#Delete", func() {
		BeforeEach(func() {
			fixturesFileNames := []string{"SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service_Delete.json",
				"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",
				"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
				"SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service_getObject_None.json",
				"SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service_getObject_None.json",
				"SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service_getObject_None.json",
				"SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service_getObject_None.json",
				"SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service_getObject_None.json",
				"SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service_getObject_None.json",
				"SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service_getObject_None.json",
				"SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service_getObject_None.json"}

			testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fixturesFileNames)
		})

		Context("when stemcell exists", func() {
			It("deletes the stemcell in collection directory that contains unpacked stemcell", func() {
				err := stemcell.Delete()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when stemcell does not exist", func() {
			BeforeEach(func() {
				fakeSoftLayerClient.FakeHttpClient.DoRawHttpRequestInt = 404
			})

			It("returns error if deleting stemcell does not exist", func() {
				err := stemcell.Delete()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
