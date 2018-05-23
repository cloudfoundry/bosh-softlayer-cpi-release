package instance_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
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
		logger = cpiLog.NewLogger(boshlog.LevelNone, "")
		virtualGuestService = NewSoftLayerVirtualGuestService(cli, uuidGen, logger)
	})

	Describe("Call CreateSshKey", func() {
		var (
			label       string
			sshKey      string
			fingerPrint string
		)

		BeforeEach(func() {
			label = "fake-sshkey"
			sshKey = "cdj33ndDf&4"
			fingerPrint = "6530389635564f6464e8e3a47d593e19"

			cli.CreateSshKeyReturns(
				&datatypes.Security_Ssh_Key{
					Id: sl.Int(32345678),
				},
				nil,
			)
		})

		It("Configure networks successfully", func() {
			_, err := virtualGuestService.CreateSshKey(label, sshKey, fingerPrint)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.CreateSshKeyCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient ConfigureNetworks call returns an error", func() {
			cli.CreateSshKeyReturns(
				&datatypes.Security_Ssh_Key{},
				errors.New("fake-client-error"),
			)

			_, err := virtualGuestService.CreateSshKey(label, sshKey, fingerPrint)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.CreateSshKeyCallCount()).To(Equal(1))
		})
	})

	Describe("Call DeleteSshKey", func() {
		var (
			sshKeyId int
		)

		BeforeEach(func() {
			sshKeyId = 52345678

			cli.DeleteSshKeyReturns(
				true,
				nil,
			)
		})

		It("Configure networks successfully", func() {
			err := virtualGuestService.DeleteSshKey(sshKeyId)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.DeleteSshKeyCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient ConfigureNetworks call returns an error", func() {
			cli.DeleteSshKeyReturns(
				false,
				errors.New("fake-client-error"),
			)

			err := virtualGuestService.DeleteSshKey(sshKeyId)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.DeleteSshKeyCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient ConfigureNetworks call returns negative result", func() {
			cli.DeleteSshKeyReturns(
				false,
				nil,
			)

			err := virtualGuestService.DeleteSshKey(sshKeyId)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to delete Ssh Public Key with id"))
			Expect(cli.DeleteSshKeyCallCount()).To(Equal(1))
		})
	})
})
