package fakes

import (
	wrdn "github.com/cloudfoundry-incubator/garden/warden"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type FakeAgentEnvServiceFactory struct {
	NewContainer       wrdn.Container
	NewAgentEnvService *FakeAgentEnvService
}

func (f *FakeAgentEnvServiceFactory) New(container wrdn.Container) bslcvm.AgentEnvService {
	f.NewContainer = container

	if f.NewAgentEnvService == nil {
		// Always return non-nil service for convenience
		return &FakeAgentEnvService{}
	}

	return f.NewAgentEnvService
}
