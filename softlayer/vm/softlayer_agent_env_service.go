package vm

import (
	"errors"
	
	boshlog "bosh/logger"
)

const (
	softLayerAgentEnvServiceLogTag = "softLayerAgentEnvService"

	softLayerAgentEnvServiceSettingsFileName  = "softlayer-cpi-agent-env.json"
	softLayerAgentEnvServiceTmpSettingsPath   = "/tmp/" + softLayerAgentEnvServiceSettingsFileName
	softLayerAgentEnvServiceFinalSettingsPath = "/var/vcap/bosh/" + softLayerAgentEnvServiceSettingsFileName
)

type SoftLayerAgentEnvService struct {
	logger    boshlog.Logger
}

func NewSoftLayerAgentEnvService(
	logger boshlog.Logger,
) SoftLayerAgentEnvService {
	return SoftLayerAgentEnvService{
		logger:    logger,
	}
}

func (s SoftLayerAgentEnvService) Fetch() (AgentEnv, error) {
	return AgentEnv{}, errors.New("Implement me!")
}

func (s SoftLayerAgentEnvService) Update(agentEnv AgentEnv) error {
	return errors.New("Implement me!")
}
