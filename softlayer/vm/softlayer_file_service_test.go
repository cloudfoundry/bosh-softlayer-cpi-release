package vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"
	fakesutil "github.com/cloudfoundry/bosh-softlayer-cpi/util/fakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	fakevm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm/fakes"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

var _ = Describe("SoftlayerFileService", func() {
	var (
		logger               boshlog.Logger
		softLayerClient      *fakeslclient.FakeSoftLayerClient
		sshClient            *fakesutil.FakeSshClient
		fs                   *fakesys.FakeFileSystem
		uuidGenerator        *fakeuuid.FakeGenerator
		vm                   *fakevm.FakeVM
		softlayerFileService SoftlayerFileService
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		sshClient = &fakesutil.FakeSshClient{}
		uuidGenerator = fakeuuid.NewFakeGenerator()
		fs = fakesys.NewFakeFileSystem()
		vm = fakevm.NewFakeVM(1234567)

		testhelpers.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Virtual_Guest_Service_getObject.json")
	})

	Describe("Upload", func() {
		It("file contents into /var/vcap/file.ext", func() {
			softlayerFileService = NewSoftlayerFileService(sshClient, logger, uuidGenerator, fs)
		        softlayerFileService.SetVM(vm)
			err := softlayerFileService.Upload("/var/vcap/file.ext", []byte("fake-contents"))7
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Download", func() {
		It("copies agent env into temporary location", func() {
			softlayerFileService = NewSoftlayerFileService(sshClient, logger, uuidGenerator, fs)
		        softlayerFileService.SetVM(vm)
			_, err := softlayerFileService.Download("/fake-download-path/file.ext")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("File not found"))
		})
	})
})
