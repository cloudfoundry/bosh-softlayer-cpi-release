package instance_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	. "bosh-softlayer-cpi/softlayer/virtual_guest_service"
	"github.com/maximilien/softlayer-go/data_types"
)

var _ = Describe("SoftLayer_Virtual_Guest_Disk", func() {
	var (
		softLayerClient     *fakeslclient.FakeClient
		uuidGen             *fakeuuid.FakeGenerator
		logger              boshlog.Logger
		virtualGuestService SoftlayerVirtualGuestService
	)

	BeforeEach(func() {
		softLayerClient = &fakeslclient.FakeClient{}
		uuidGen = &fakeuuid.FakeGenerator{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		virtualGuestService = NewSoftLayerVirtualGuestService(softLayerClient, uuidGen, logger)
	})

	Describe("AttachEphemeralDisk", func() {
		var (
			vmID     int
			diskSize int
		)

		BeforeEach(func() {
			vmID = 12345678
			diskSize = 1024
		})

		It("Attach successfully", func() {
			softLayerClient.AttachSecondDiskToInstanceReturns(
				nil,
			)

			err := virtualGuestService.AttachEphemeralDisk(vmID, diskSize)
			Expect(err).NotTo(HaveOccurred())
			Expect(softLayerClient.AttachSecondDiskToInstanceCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient attach second disk to instance call returns an error", func() {
			softLayerClient.AttachSecondDiskToInstanceReturns(
				errors.New("fake-client-error"),
			)

			err := virtualGuestService.AttachEphemeralDisk(vmID, diskSize)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(softLayerClient.AttachSecondDiskToInstanceCallCount()).To(Equal(1))
		})
	})

	Describe("AttachDisk", func() {
		var (
			vmID     int
			diskSize int
		)

		BeforeEach(func() {
			vmID = 12345678
			diskSize = 1024

			softLayerClient.GetNetworkStorageTargetReturns(
				"fake-target-address",
				true,
				nil,
			)
			softLayerClient.GetInstanceReturns(
				*data_types.Virtual_Guest{},
				true,
				nil,
			)

		})

		It("Attach successfully", func() {


			err := virtualGuestService.AttachDisk(vmID, diskSize)

		})
	})
})
