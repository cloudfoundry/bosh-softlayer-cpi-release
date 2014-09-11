package vm

import (
	boshlog "bosh/logger"
	wrdn "github.com/cloudfoundry-incubator/garden/warden"
)

type SoftLayerAgentEnvServiceFactory struct {
	logger boshlog.Logger
}

func NewSoftLayerAgentEnvServiceFactory(logger boshlog.Logger) SoftLayerAgentEnvServiceFactory {
	return SoftLayerAgentEnvServiceFactory{logger: logger}
}

func (f SoftLayerAgentEnvServiceFactory) New(container wrdn.Container) AgentEnvService {
	return NewSoftLayerAgentEnvService(container, f.logger)
}
