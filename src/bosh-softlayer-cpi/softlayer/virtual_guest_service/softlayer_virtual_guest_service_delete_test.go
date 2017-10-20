package instance_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	cpiLog "bosh-softlayer-cpi/logger"
	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	. "bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

var _ = Describe("Virtual Guest Service", func() {
	var (
		cli                 *fakeslclient.FakeClient
		uuidGen             *fakeuuid.FakeGenerator
		logger              cpiLog.Logger
		virtualGuestService SoftlayerVirtualGuestService
	)

	BeforeEach(func() {
		cli = &fakeslclient.FakeClient{}
		uuidGen = &fakeuuid.FakeGenerator{}
		logger = cpiLog.NewLogger(boshlog.LevelDebug, "")
		virtualGuestService = NewSoftLayerVirtualGuestService(cli, uuidGen, logger)
	})

	Describe("Call Delete", func() {
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
				cli.DeleteInstanceFromVPSReturns(
					nil,
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.DeleteInstanceFromVPSCallCount()).To(Equal(1))
				Expect(cli.CancelInstanceCallCount()).To(Equal(0))
			})

			It("Return error if softLayerClient delete instance from VPS call returns an error", func() {
				cli.DeleteInstanceFromVPSReturns(
					errors.New("fake-client-error"),
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.DeleteInstanceFromVPSCallCount()).To(Equal(1))
				Expect(cli.CancelInstanceCallCount()).To(Equal(0))
			})
		})

		Context("Clean up from virtual VirtualGuestService", func() {

			BeforeEach(func() {
				enableVps = false
			})

			It("Clean up successfully", func() {
				cli.CancelInstanceReturns(
					nil,
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.DeleteInstanceFromVPSCallCount()).To(Equal(0))
				Expect(cli.CancelInstanceCallCount()).To(Equal(1))
			})

			It("Return error if softLayerClient delete instance from VPS call returns an error", func() {
				cli.CancelInstanceReturns(
					errors.New("fake-client-error"),
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.DeleteInstanceFromVPSCallCount()).To(Equal(0))
				Expect(cli.CancelInstanceCallCount()).To(Equal(1))
			})
		})
	})
})
