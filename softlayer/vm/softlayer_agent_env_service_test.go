package vm_test

import (
	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("SoftLayerAgentEnvService", func() {
	var (
		vmId            int
		softLayerClient *fakeslclient.FakeSoftLayerClient
		agentEnvService SoftLayerAgentEnvService
		logger          boshlog.Logger
	)

	BeforeEach(func() {
		vmId = 1234567
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		logger = boshlog.NewLogger(boshlog.LevelNone)
		agentEnvService = NewSoftLayerAgentEnvService(vmId, softLayerClient, logger)
	})

	Context("#Fetch", func() {
		It("Returns an AgentEnv object built with current metadata when fetched", func() {
			//Implement me!
		})
	})

	Context("#Update", func() {
		It("Sets the VM's metadata using the AgentEnv object passed", func() {
			//Implement me!
		})
	})
})
