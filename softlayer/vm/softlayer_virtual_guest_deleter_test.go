package vm_test

import (
	"time"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"
	slh "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
)

var _ = Describe("SoftlayerVirtualGuestDeleter", func() {
	var (
		fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient
		logger          boshlog.Logger
		deleter         VMDeleter
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
		})

		Context("when deleting virtual guest succeeds", func() {
			BeforeEach(func() {
				setFakeSoftlayerClientDeleteObjectTestFixtures(fakeSoftLayerClient)
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

		Context("when deleting object fails", func() {
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

func setFakeSoftlayerClientDeleteObjectTestFixtures(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_deleteObject_true.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}