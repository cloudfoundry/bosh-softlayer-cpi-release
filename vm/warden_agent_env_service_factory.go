package vm

import (
	boshlog "bosh/logger"
	wrdn "github.com/cloudfoundry-incubator/garden/warden"
)

type WardenAgentEnvServiceFactory struct {
	logger boshlog.Logger
}

func NewWardenAgentEnvServiceFactory(logger boshlog.Logger) WardenAgentEnvServiceFactory {
	return WardenAgentEnvServiceFactory{logger: logger}
}

func (f WardenAgentEnvServiceFactory) New(container wrdn.Container) AgentEnvService {
	return NewWardenAgentEnvService(container, f.logger)
}
