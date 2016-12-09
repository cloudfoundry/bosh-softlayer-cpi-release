package vm_test

import (
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	slh "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("SoftlayerVirtualGuestDeleter", func() {
	var (
		fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient
		logger              boshlog.Logger
		deleter             VMDeleter
	)

	BeforeEach(func() {
		fakeSoftLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		logger = boshlog.NewLogger(boshlog.LevelNone)
		deleter = NewSoftLayerVMDeleter(fakeSoftLayerClient, logger)
		slh.TIMEOUT = 2 * time.Second
		slh.POLLING_INTERVAL = 1 * time.Second
	})

	Describe("Delete", func() {
		var (
			err error
		)

		JustBeforeEach(func() {
			err = deleter.Delete(1234567)
			fakeSoftLayerClient.FakeHttpClient.DoRawHttpRequestResponse = []byte("true")
		})

		Context("when deleting virtual guest succeeds", func() {
			BeforeEach(func() {
				setFakeSoftlayerClientDeleteObjectTrueTestFixtures(fakeSoftLayerClient)
			})

			It("returns no error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when virtual guest have runing sections", func() {
			BeforeEach(func() {
				testhelpers.SetTestFixtureForFakeSoftLayerClient(fakeSoftLayerClient, "SoftLayer_Virtual_Guest_Service_getActiveTransactions.json")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("have no pending transactions before deleting vm"))
			})
		})

		Context("when deleting object and error occures", func() {
			BeforeEach(func() {
				testhelpers.SetTestFixtureForFakeSoftLayerClient(fakeSoftLayerClient, "SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Deleting SoftLayer VirtualGuest from client"))
			})
		})
	})

})

func setFakeSoftlayerClientDeleteObjectTrueTestFixtures(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_deleteObject_true.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}
