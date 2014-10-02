package vm_test

import (
	. "github.com/onsi/ginkgo"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("SoftLayerCreator", func() {
	var (
		softLayerClient        *fakeslclient.FakeSoftLayerClient
		agentEnvServiceFactory *fakevm.FakeAgentEnvServiceFactory
		agentOptions           AgentOptions
		logger                 boshlog.Logger
		creator                SoftLayerCreator
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		agentEnvServiceFactory = &fakevm.FakeAgentEnvServiceFactory{}
		agentOptions = AgentOptions{Mbus: "fake-mbus"}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		creator = NewSoftLayerCreator(
			softLayerClient,
			agentEnvServiceFactory,
			agentOptions,
			logger,
		)
	})

	Describe("Create", func() {
	})
})
