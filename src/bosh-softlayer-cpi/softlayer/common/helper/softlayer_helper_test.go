package helper_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "bosh-softlayer-cpi/test_helpers"

	fakescommon "bosh-softlayer-cpi/softlayer/common/fakes"
	fakesutil "bosh-softlayer-cpi/util/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	slh "bosh-softlayer-cpi/softlayer/common/helper"
)

var _ = Describe("SoftLayerVirtualGuest", func() {
	var (
		fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient
		sshClient           *fakesutil.FakeSshClient
		agentEnvService     *fakescommon.FakeAgentEnvService
		logger              boshlog.Logger
	)

	BeforeEach(func() {
		fakeSoftLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		sshClient = &fakesutil.FakeSshClient{}
		agentEnvService = &fakescommon.FakeAgentEnvService{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		slh.TIMEOUT = 10 * time.Millisecond
		slh.POLLING_INTERVAL = 1 * time.Millisecond
	})

	Describe("WaitForVirtualGuestLastCompleteTransaction", func() {
		Context("where the transcation is not completed successfully", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClientbyLevels(fakeSoftLayerClient, fileNames, 3)
			})

			It("returns nil", func() {
				err := slh.WaitForVirtualGuestLastCompleteTransaction(fakeSoftLayerClient, 1234567, "Service Setup")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("where GetLastTransaction returns error", func() {
			BeforeEach(func() {
				fakeSoftLayerClient.FakeHttpClient.DoRawHttpRequestError = errors.New("Error occurred when getting the last transaction")
			})

			It("returns error", func() {
				err := slh.WaitForVirtualGuestLastCompleteTransaction(fakeSoftLayerClient, 1234567, "Service Setup")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("where the transcation is not completed within timeout duration", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_inProgress.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClientbyLevels(fakeSoftLayerClient, fileNames, 3)
			})

			It("returns error", func() {
				err := slh.WaitForVirtualGuestLastCompleteTransaction(fakeSoftLayerClient, 1234567, "Service Setup")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("AttachEphemeralDiskToVirtualGuest", func() {
		Context("where the local disk is attached successfully", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
					"SoftLayer_Virtual_Guest_Service_getUpgradeItemPrices.json",
					"SoftLayer_Virtual_Guest_Service_getLocalDiskFlag_local.json",
					"SoftLayer_Product_Order_placeOrder.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_CloudInstanceUpgrade.json",
					"SoftLayer_Virtual_Guest_Service_getPowerState.json",
					"SoftLayer_Virtual_Guest_Service_getBlockDevices.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClientbyLevels(fakeSoftLayerClient, fileNames, 3)
			})

			It("returns nil", func() {
				err := slh.AttachEphemeralDiskToVirtualGuest(fakeSoftLayerClient, 12345, 25, logger)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("where the local disk is not attached properly", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
					"SoftLayer_Virtual_Guest_Service_getUpgradeItemPrices.json",
					"SoftLayer_Virtual_Guest_Service_getLocalDiskFlag_local.json",
					"SoftLayer_Product_Order_placeOrder.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
					"SoftLayer_Virtual_Guest_Service_getLastTransaction_CloudInstanceUpgrade.json",
					"SoftLayer_Virtual_Guest_Service_getPowerState.json",
					"SoftLayer_Virtual_Guest_Service_getBlockDevices_2_devices.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClientbyLevels(fakeSoftLayerClient, fileNames, 3)
			})

			It("returns error", func() {
				err := slh.AttachEphemeralDiskToVirtualGuest(fakeSoftLayerClient, 12345, 25, logger)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
