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

var _ = Describe("Virtual Guest Service", func() {
	var (
		cli                 *fakeslclient.FakeClient
		uuidGen             *fakeuuid.FakeGenerator
		logger              boshlog.Logger
		virtualGuestService SoftlayerVirtualGuestService
	)

	BeforeEach(func() {
		cli = &fakeslclient.FakeClient{}
		uuidGen = &fakeuuid.FakeGenerator{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		virtualGuestService = NewSoftLayerVirtualGuestService(cli, uuidGen, logger)
	})

	Describe("Call Reboot", func() {
		var (
			vmID int
		)

		BeforeEach(func() {
			vmID = 12345678

			cli.RebootInstanceReturns(
				nil,
			)
		})

		It("Reboot instance successfully", func() {
			err := virtualGuestService.Reboot(vmID)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.RebootInstanceCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient RebootInstance call returns an error", func() {
			cli.RebootInstanceReturns(
				errors.New("fake-client-error"),
			)

			err := virtualGuestService.Reboot(vmID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.RebootInstanceCallCount()).To(Equal(1))
		})
	})
})
