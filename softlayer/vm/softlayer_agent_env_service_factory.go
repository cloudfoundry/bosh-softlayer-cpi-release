package vm

import (
	boshlog "bosh/logger"
)

type SoftLayerAgentEnvServiceFactory struct {
	logger boshlog.Logger
}

func NewSoftLayerAgentEnvServiceFactory(logger boshlog.Logger) SoftLayerAgentEnvServiceFactory {
	return SoftLayerAgentEnvServiceFactory{logger: logger}
}

func (f SoftLayerAgentEnvServiceFactory) New() AgentEnvService {
	return NewSoftLayerAgentEnvService(f.logger)
}
