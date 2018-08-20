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

	Describe("Call UpgradeInstance", func() {
		var (
			vmID          int
			cpu           int
			memory        int
			network       int
			privateCPU    bool
			dedicatedHost bool
		)

		BeforeEach(func() {
			vmID = 12345678
			cpu = 4
			memory = 2048
			network = 1000
			privateCPU = false
			dedicatedHost = false

			cli.UpgradeInstanceConfigReturns(
				nil,
			)
		})

		It("Configure networks successfully", func() {
			err := virtualGuestService.UpgradeInstance(vmID, cpu, memory, network, privateCPU, dedicatedHost)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.UpgradeInstanceConfigCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient ConfigureNetworks call returns an error", func() {
			cli.UpgradeInstanceConfigReturns(
				errors.New("fake-client-error"),
			)

			err := virtualGuestService.UpgradeInstance(vmID, cpu, memory, network, privateCPU, dedicatedHost)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.UpgradeInstanceConfigCallCount()).To(Equal(1))
		})
	})
})
