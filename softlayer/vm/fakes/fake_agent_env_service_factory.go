package fakes

import (
	bslcpi "github.com/maximilien/bosh-softlayer-cpi/softlayer/cpi"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type FakeAgentEnvServiceFactory struct {
	NewContainer       bslcpi.Container
	NewAgentEnvService *FakeAgentEnvService
}

func (f *FakeAgentEnvServiceFactory) New(container bslcpi.Container) bslcvm.AgentEnvService {
	f.NewContainer = container

	if f.NewAgentEnvService == nil {
		// Always return non-nil service for convenience
		return &FakeAgentEnvService{}
	}

	return f.NewAgentEnvService
}
