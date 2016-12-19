package common

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"strconv"
)

type SoftLayerAgentEnvServiceFactory struct {
	registryOptions RegistryOptions
	logger          boshlog.Logger
}

func NewSoftLayerAgentEnvServiceFactory(
	registryOptions RegistryOptions,
	logger boshlog.Logger,
) SoftLayerAgentEnvServiceFactory {
	return SoftLayerAgentEnvServiceFactory{
		registryOptions: registryOptions,
		logger:          logger,
	}
}

func (f SoftLayerAgentEnvServiceFactory) New(
	vm VM,
	softlayerFileService SoftlayerFileService,
) AgentEnvService {
	if len(f.registryOptions.Host) > 0 {
		return NewRegistryAgentEnvService(f.registryOptions, strconv.Itoa(vm.ID()), f.logger)
	}
	return NewFSAgentEnvService(vm, softlayerFileService, f.logger)
}
