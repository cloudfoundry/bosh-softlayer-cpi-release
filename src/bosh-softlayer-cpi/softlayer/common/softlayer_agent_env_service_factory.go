package common

import (
	"fmt"
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
		endpoint := fmt.Sprintf(
			"http://%s:%s@%s:%d",
			f.registryOptions.Username,
			f.registryOptions.Password,
			f.registryOptions.Host,
			f.registryOptions.Port,
		)
		return NewRegistryAgentEnvService(endpoint, strconv.Itoa(*vm.ID()), f.logger)
	}
	return NewFSAgentEnvService(vm, softlayerFileService, f.logger)
}
