package vm

import (
	sl "github.com/maximilien/softlayer-go/softlayer"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type SoftLayerAgentEnvServiceFactory struct {
	client sl.Client
	logger boshlog.Logger
}

func NewSoftLayerAgentEnvServiceFactory(client sl.Client, logger boshlog.Logger) SoftLayerAgentEnvServiceFactory {
	return SoftLayerAgentEnvServiceFactory{
		client: client,
		logger: logger,
	}
}

func (f SoftLayerAgentEnvServiceFactory) New() AgentEnvService {
	return NewSoftLayerAgentEnvService(f.client, f.logger)
}
