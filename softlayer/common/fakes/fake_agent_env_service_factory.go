package fakes

import (
	bslvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
)

type FakeAgentEnvServiceFactory struct {
	NewAgentEnvService *FakeAgentEnvService
}

func (f *FakeAgentEnvServiceFactory) New(
	vm bslvm.VM,
	softlayerFileService bslvm.SoftlayerFileService,
) bslvm.AgentEnvService {
	if f.NewAgentEnvService == nil {
		// Always return non-nil service for convenience
		return &FakeAgentEnvService{}
	}

	return f.NewAgentEnvService
}
