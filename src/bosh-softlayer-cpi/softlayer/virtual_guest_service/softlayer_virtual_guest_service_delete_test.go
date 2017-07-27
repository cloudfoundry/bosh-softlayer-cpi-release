package instance_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	. "bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

var _ = Describe("SoftLayer_Virtual_Guest_Delete", func() {
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

	Describe("Delete", func() {
		var (
			vmID      int
			enableVps bool
		)

		BeforeEach(func() {
			vmID = 12345678
		})

		Context("Clean up from VPS", func() {

			BeforeEach(func() {
				enableVps = true
			})

			It("Clean up successfully", func() {
				softLayerClient.DeleteInstanceFromVPSReturns(
					nil,
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).NotTo(HaveOccurred())
				Expect(softLayerClient.DeleteInstanceFromVPSCallCount()).To(Equal(1))
				Expect(softLayerClient.CancelInstanceCallCount()).To(Equal(0))
			})

			It("Return error if softLayerClient delete instance from VPS call returns an error", func() {
				softLayerClient.DeleteInstanceFromVPSReturns(
					errors.New("fake-client-error"),
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(softLayerClient.DeleteInstanceFromVPSCallCount()).To(Equal(1))
				Expect(softLayerClient.CancelInstanceCallCount()).To(Equal(0))
			})
		})

		Context("Clean up from virtual VirtualGuestService", func() {

			BeforeEach(func() {
				enableVps = false
			})

			It("Clean up successfully", func() {
				softLayerClient.CancelInstanceReturns(
					nil,
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).NotTo(HaveOccurred())
				Expect(softLayerClient.DeleteInstanceFromVPSCallCount()).To(Equal(0))
				Expect(softLayerClient.CancelInstanceCallCount()).To(Equal(1))
			})

			It("Return error if softLayerClient delete instance from VPS call returns an error", func() {
				softLayerClient.CancelInstanceReturns(
					errors.New("fake-client-error"),
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(softLayerClient.DeleteInstanceFromVPSCallCount()).To(Equal(0))
				Expect(softLayerClient.CancelInstanceCallCount()).To(Equal(1))
			})
		})
	})
})
