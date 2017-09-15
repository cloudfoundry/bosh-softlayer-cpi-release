package instance_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	. "bosh-softlayer-cpi/softlayer/virtual_guest_service"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("Virtual Guest Service", func() {
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

	Describe("Call Find", func() {
		var (
			vmID int
		)

		BeforeEach(func() {
			vmID = 12345678
		})

		It("Find vm successfully", func() {
			softLayerClient.GetInstanceReturns(
				&datatypes.Virtual_Guest{
					Id: sl.Int(12345678),
				},
				true,
				nil,
			)

			vm, err := virtualGuestService.Find(vmID)
			Expect(err).NotTo(HaveOccurred())
			Expect(softLayerClient.GetInstanceCallCount()).To(Equal(1))
			Expect(*vm.Id).To(Equal(vmID))
		})

		It("Return error if softLayerClient GetInstance returns an error", func() {
			softLayerClient.GetInstanceReturns(
				&datatypes.Virtual_Guest{},
				false,
				errors.New("fake-client-error"),
			)

			_, err := virtualGuestService.Find(vmID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(softLayerClient.GetInstanceCallCount()).To(BeNumerically(">=", 2))
		})

		It("Return error if softLayerClient GetInstance returns non-existing", func() {
			softLayerClient.GetInstanceReturns(
				&datatypes.Virtual_Guest{},
				false,
				nil,
			)

			_, err := virtualGuestService.Find(vmID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(softLayerClient.GetInstanceCallCount()).To(Equal(1))
		})
	})

	Describe("Call FindByPrimaryBackendIp", func() {
		var (
			vmIP string
		)

		BeforeEach(func() {
			vmIP = "fake-vm-ip"
		})

		It("Find vm successfully", func() {
			softLayerClient.GetInstanceByPrimaryBackendIpAddressReturns(
				&datatypes.Virtual_Guest{
					Id: sl.Int(12345678),
					PrimaryBackendIpAddress: sl.String(vmIP),
				},
				true,
				nil,
			)

			vm, err := virtualGuestService.FindByPrimaryBackendIp(vmIP)
			Expect(err).NotTo(HaveOccurred())
			Expect(softLayerClient.GetInstanceByPrimaryBackendIpAddressCallCount()).To(Equal(1))
			Expect(*vm.PrimaryBackendIpAddress).To(Equal(vmIP))
		})

		It("Return error if softLayerClient GetInstanceByPrimaryBackendIpAddress returns an error", func() {
			softLayerClient.GetInstanceByPrimaryBackendIpAddressReturns(
				&datatypes.Virtual_Guest{},
				false,
				errors.New("fake-client-error"),
			)

			_, err := virtualGuestService.FindByPrimaryBackendIp(vmIP)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(softLayerClient.GetInstanceByPrimaryBackendIpAddressCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient GetInstanceByPrimaryBackendIpAddress returns non-existing", func() {
			softLayerClient.GetInstanceByPrimaryBackendIpAddressReturns(
				&datatypes.Virtual_Guest{},
				false,
				nil,
			)

			_, err := virtualGuestService.FindByPrimaryBackendIp(vmIP)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(softLayerClient.GetInstanceByPrimaryBackendIpAddressCallCount()).To(Equal(1))
		})
	})

	Describe("Call FindByPrimaryIp", func() {
		var (
			vmIP string
		)

		BeforeEach(func() {
			vmIP = "fake-vm-ip"
		})

		It("Find vm successfully", func() {
			softLayerClient.GetInstanceByPrimaryIpAddressReturns(
				&datatypes.Virtual_Guest{
					Id:               sl.Int(12345678),
					PrimaryIpAddress: sl.String(vmIP),
				},
				true,
				nil,
			)

			vm, err := virtualGuestService.FindByPrimaryIp(vmIP)
			Expect(err).NotTo(HaveOccurred())
			Expect(softLayerClient.GetInstanceByPrimaryIpAddressCallCount()).To(Equal(1))
			Expect(*vm.PrimaryIpAddress).To(Equal(vmIP))
		})

		It("Return error if softLayerClient GetInstanceByPrimaryIpAddress returns an error", func() {
			softLayerClient.GetInstanceByPrimaryIpAddressReturns(
				&datatypes.Virtual_Guest{},
				false,
				errors.New("fake-client-error"),
			)

			_, err := virtualGuestService.FindByPrimaryIp(vmIP)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(softLayerClient.GetInstanceByPrimaryIpAddressCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient GetInstanceByPrimaryIpAddress returns non-exist", func() {
			softLayerClient.GetInstanceByPrimaryIpAddressReturns(
				&datatypes.Virtual_Guest{},
				false,
				nil,
			)

			_, err := virtualGuestService.FindByPrimaryIp(vmIP)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(softLayerClient.GetInstanceByPrimaryIpAddressCallCount()).To(Equal(1))
		})
	})
})
