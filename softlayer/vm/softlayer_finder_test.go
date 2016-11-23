package vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	fakebmsclient "github.com/cloudfoundry-community/bosh-softlayer-tools/clients/fakes"
	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("SoftLayerFinder", func() {
	var (
		softLayerClient        *fakeslclient.FakeSoftLayerClient
		baremetalClient        *fakebmsclient.FakeBmpClient
		agentEnvServiceFactory *fakescommon.FakeAgentEnvServiceFactory
		logger                 boshlog.Logger
		finder                 VMFinder
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		baremetalClient = fakebmsclient.NewFakeBmpClient("fake-username", "fake-api-key", "fake-url", "fake-configpath")
		agentEnvServiceFactory = &fakescommon.FakeAgentEnvServiceFactory{}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		finder = NewSoftLayerFinder(
			softLayerClient,
			baremetalClient,
			agentEnvServiceFactory,
			logger,
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
