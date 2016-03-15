package fakes

import (
	bslvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type FakeAgentEnvServiceFactory struct {
	NewAgentEnvService *FakeAgentEnvService
}

func (f *FakeAgentEnvServiceFactory) New(
	softlayerFileService bslvm.SoftlayerFileService,
	instanceID string,
) bslvm.AgentEnvService {
	if f.NewAgentEnvService == nil {
		// Always return non-nil service for convenience
		return &FakeAgentEnvService{}
	}

	return f.NewAgentEnvService
}
