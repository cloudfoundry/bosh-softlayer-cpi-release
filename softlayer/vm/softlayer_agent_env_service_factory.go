package vm

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type SoftLayerAgentEnvServiceFactory struct {
	agentEnvService string
	registryOptions RegistryOptions
	logger          boshlog.Logger
}

func NewSoftLayerAgentEnvServiceFactory(
	agentEnvService string,
	registryOptions RegistryOptions,
	logger boshlog.Logger,
) SoftLayerAgentEnvServiceFactory {
	return SoftLayerAgentEnvServiceFactory{
		logger:          logger,
		agentEnvService: agentEnvService,
		registryOptions: registryOptions,
	}
}

func (f SoftLayerAgentEnvServiceFactory) New(
	vm VM,
	softlayerFileService SoftlayerFileService,
) AgentEnvService {
	return NewFSAgentEnvService(vm, softlayerFileService, f.logger)
}
