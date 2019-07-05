package instance_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/bluebosh/bosh-utils/logger"
	fakeuuid "github.com/bluebosh/bosh-utils/uuid/fakes"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"

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
				cli.GetInstanceReturns(
					&datatypes.Virtual_Guest{
						Id: sl.Int(vmID),
					},
					true,
					nil,
				)
				cli.DeleteInstanceFromVPSReturns(
					nil,
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.GetInstanceCallCount()).To(Equal(1))
				Expect(cli.DeleteInstanceFromVPSCallCount()).To(Equal(1))
				Expect(cli.CancelInstanceCallCount()).To(Equal(0))
			})

			It("Return error if softLayerClient delete instance from VPS call returns an error", func() {
				cli.GetInstanceReturns(
					&datatypes.Virtual_Guest{
						Id: sl.Int(vmID),
					},
					true,
					nil,
				)
				cli.DeleteInstanceFromVPSReturns(
					errors.New("fake-client-error"),
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.DeleteInstanceFromVPSCallCount()).To(Equal(1))
				Expect(cli.CancelInstanceCallCount()).To(Equal(0))
			})

			It("Return nil when softLayerClient find instance call return an ObjectNotFoundError error", func() {
				cli.GetInstanceReturns(
					&datatypes.Virtual_Guest{},
					false,
					nil,
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.GetInstanceCallCount()).To(Equal(1))
				Expect(cli.DeleteInstanceFromVPSCallCount()).To(Equal(0))
				Expect(cli.CancelInstanceCallCount()).To(Equal(0))
			})
		})

		Context("Clean up from virtual VirtualGuestService", func() {

			BeforeEach(func() {
				enableVps = false
			})

			It("Clean up successfully", func() {
				cli.GetInstanceReturns(
					&datatypes.Virtual_Guest{
						Id: sl.Int(vmID),
					},
					true,
					nil,
				)
				cli.CancelInstanceReturns(
					nil,
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.GetInstanceCallCount()).To(Equal(1))
				Expect(cli.DeleteInstanceFromVPSCallCount()).To(Equal(0))
				Expect(cli.CancelInstanceCallCount()).To(Equal(1))
			})

			It("Return error if softLayerClient delete instance from VPS call returns an error", func() {
				cli.GetInstanceReturns(
					&datatypes.Virtual_Guest{
						Id: sl.Int(vmID),
					},
					true,
					nil,
				)
				cli.CancelInstanceReturns(
					errors.New("fake-client-error"),
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).To(HaveOccurred())
				Expect(cli.GetInstanceCallCount()).To(Equal(1))
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.DeleteInstanceFromVPSCallCount()).To(Equal(0))
				Expect(cli.CancelInstanceCallCount()).To(Equal(1))
			})

			It("Return nil when softLayerClient find instance call return an ObjectNotFoundError error", func() {
				cli.GetInstanceReturns(
					&datatypes.Virtual_Guest{},
					false,
					nil,
				)

				err := virtualGuestService.Delete(vmID, enableVps)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.GetInstanceCallCount()).To(Equal(1))
				Expect(cli.DeleteInstanceFromVPSCallCount()).To(Equal(0))
				Expect(cli.CancelInstanceCallCount()).To(Equal(0))
			})
		})
	})
})
