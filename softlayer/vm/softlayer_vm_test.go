package vm_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	boshlog "bosh/logger"

	fakebslclient "github.com/maximilien/softlayer-go/client/fakes"

	fakedisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk/fakes"
	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
)

var _ = Describe("SoftLayerVM", func() {
	var (
		softLayerClient *fakebslclient.FakeClient
		agentEnvService *fakevm.FakeAgentEnvService
		logger          boshlog.Logger
		vm              SoftLayerVM
	)

	BeforeEach(func() {
		softLayerClient = fakebslcpi.New()
		agentEnvService = &fakevm.FakeAgentEnvService{}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		vm = NewSoftLayerVM(
			"fake-vm-id",
			softLayerClient,
			agentEnvService,
			logger,
		)
	})

	Describe("Delete", func() {
	})

	Describe("AttachDisk", func() {
	})

	Describe("DetachDisk", func() {
	})
})
