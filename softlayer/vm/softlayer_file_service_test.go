package vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
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
		fs                   *fakesys.FakeFileSystem
		uuidGenerator        *fakeuuid.FakeGenerator
		softlayerFileService SoftlayerFileService
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		sshClient = fakesutil.NewFakeSshClient()
		uuidGenerator = fakeuuid.NewFakeGenerator()
		fs = fakesys.NewFakeFileSystem()

		testhelpers.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Virtual_Guest_Service_getObject.json")
	})

	Describe("Upload", func() {
		It("file contents into /var/vcap/file.ext", func() {
			expectedCmdResults := []string{
				"",
			}
			testhelpers.SetTestFixturesForFakeSSHClient(sshClient, expectedCmdResults, nil)

			var virtualGuestService softlayer.SoftLayer_Virtual_Guest_Service
			virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
			Expect(err).ToNot(HaveOccurred())
			virtualGuest, err := virtualGuestService.GetObject(1234567)
			Expect(err).ToNot(HaveOccurred())

			softlayerFileService = NewSoftlayerFileService(sshClient, virtualGuest, logger, uuidGenerator, fs)
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

			softlayerFileService = NewSoftlayerFileService(sshClient, virtualGuest, logger, uuidGenerator, fs)
			_, err = softlayerFileService.Download("/fake-download-path/file.ext")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("File not found"))
		})

	})
})
