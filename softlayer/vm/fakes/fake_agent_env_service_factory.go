package fakes

import (
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type FakeAgentEnvServiceFactory struct {
	NewAgentEnvService *FakeAgentEnvService
}

func (f *FakeAgentEnvServiceFactory) New() bslcvm.AgentEnvService {
	if f.NewAgentEnvService == nil {
		// Always return non-nil service for convenience
		return &FakeAgentEnvService{}
	}

	return f.NewAgentEnvService
}
