package vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	fakevm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm/fakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("SoftLayerFinder", func() {
	var (
		softLayerClient        *fakeslclient.FakeSoftLayerClient
		agentEnvServiceFactory *fakevm.FakeAgentEnvServiceFactory
		fs                     *fakesys.FakeFileSystem
		uuidGenerator          *fakeuuid.FakeGenerator
		logger                 boshlog.Logger
		finder                 SoftLayerFinder
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		uuidGenerator = fakeuuid.NewFakeGenerator()
		fs = fakesys.NewFakeFileSystem()
		agentEnvServiceFactory = &fakevm.FakeAgentEnvServiceFactory{}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		finder = NewSoftLayerFinder(
			softLayerClient,
			agentEnvServiceFactory,
			logger,
			uuidGenerator,
			fs,
		)
	})

	Describe("Find", func() {
		var (
			vmID int
		)

		Context("when the VM ID is valid and existing", func() {
			BeforeEach(func() {
				vmID = 1234567
				testhelpers.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Virtual_Guest_Service_getObject.json")
			})

			It("finds and returns a new SoftLayerVM object with correct ID", func() {
				vm, found, err := finder.Find(vmID)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue(), "found VM")
				Expect(vm.ID()).To(Equal(vmID), "VM ID match")
			})
		})

	})
})
