package vm

import (
	boshlog "bosh/logger"

	bslcpi "github.com/maximilien/bosh-softlayer-cpi/softlayer/cpi"
)

type SoftLayerAgentEnvServiceFactory struct {
	logger boshlog.Logger
}

func NewSoftLayerAgentEnvServiceFactory(logger boshlog.Logger) SoftLayerAgentEnvServiceFactory {
	return SoftLayerAgentEnvServiceFactory{logger: logger}
}

func (f SoftLayerAgentEnvServiceFactory) New(container bslcpi.Container) AgentEnvService {
	return NewSoftLayerAgentEnvService(container, f.logger)
}
