package vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "github.com/maximilien/bosh-softlayer-cpi/test_helpers"
	fakesutil "github.com/maximilien/bosh-softlayer-cpi/util/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	softlayer "github.com/maximilien/softlayer-go/softlayer"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

var _ = Describe("SoftlayerFileService", func() {
	var (
		logger               boshlog.Logger
		softLayerClient      *fakeslclient.FakeSoftLayerClient
		sshClient            *fakesutil.FakeSshClient
		softlayerFileService SoftlayerFileService
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		sshClient = fakesutil.NewFakeSshClient()

		testhelpers.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Virtual_Guest_Service_getObject.json")
	})

	Describe("Upload", func() {
		It("scp content into vm at /tmp/destination", func() {
			expectedCmdResults := []string{
				"",
			}
			testhelpers.SetTestFixturesForFakeSSHClient(sshClient, expectedCmdResults, nil)

			var virtualGuestService softlayer.SoftLayer_Virtual_Guest_Service
			virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
			Expect(err).ToNot(HaveOccurred())
			virtualGuest, err := virtualGuestService.GetObject(1234567)
			Expect(err).ToNot(HaveOccurred())

			softlayerFileService = NewSoftlayerFileService(sshClient, virtualGuest, logger)
			err = softlayerFileService.Upload("/var/vcap/file.ext", []byte("fake-contents"))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Download", func() {
		It("copies agent env into temporary location", func() {
			expectedCmdResults := []string{
				"",
			}
			testhelpers.SetTestFixturesForFakeSSHClient(sshClient, expectedCmdResults, nil)

			var virtualGuestService softlayer.SoftLayer_Virtual_Guest_Service
			virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
			Expect(err).ToNot(HaveOccurred())
			virtualGuest, err := virtualGuestService.GetObject(1234567)
			Expect(err).ToNot(HaveOccurred())

			softlayerFileService = NewSoftlayerFileService(sshClient, virtualGuest, logger)
			_, err = softlayerFileService.Download("/fake-download-path/file.ext")
			Expect(err).ToNot(HaveOccurred())
		})

	})
})
