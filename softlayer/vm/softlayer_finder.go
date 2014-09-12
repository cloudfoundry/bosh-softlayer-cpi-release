package vm

import (
	boshlog "bosh/logger"

	sl "github.com/maximilien/softlayer-go/softlayer"
)

const softLayerFinderLogTag = "SoftLayerFinder"

type SoftLayerFinder struct {
	softLayerClient        sl.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	logger boshlog.Logger
}

func NewSoftLayerFinder(
	softLayerClient sl.Client,
	agentEnvServiceFactory AgentEnvServiceFactory,
	logger boshlog.Logger,
) SoftLayerFinder {
	return SoftLayerFinder{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,

		logger: logger,
	}
}

func (f SoftLayerFinder) Find(id int) (VM, bool, error) {
	f.logger.Debug(softLayerFinderLogTag, "Finding container with ID '%s'", id)

	//Find VM here using SL client

	f.logger.Debug(softLayerFinderLogTag, "Did not find container with ID '%s'", id)

	return nil, false, nil
}
