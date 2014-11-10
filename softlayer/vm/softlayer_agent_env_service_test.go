package vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("SoftLayerAgentEnvService", func() {
	var (
		softLayerClient *fakeslclient.FakeSoftLayerClient
		agentEnvService SoftLayerAgentEnvService
		logger          boshlog.Logger
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		logger = boshlog.NewLogger(boshlog.LevelNone)
		agentEnvService = NewSoftLayerAgentEnvService(softLayerClient, logger)
	})
	})
})
