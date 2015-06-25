package vm

import (
	"errors"

	sl "github.com/maximilien/softlayer-go/softlayer"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const (
	softLayerAgentEnvServiceLogTag = "softLayerAgentEnvService"

	softLayerAgentEnvServiceSettingsFileName  = "softlayer-cpi-agent-env.json"
	softLayerAgentEnvServiceTmpSettingsPath   = "/tmp/" + softLayerAgentEnvServiceSettingsFileName
	softLayerAgentEnvServiceFinalSettingsPath = "/var/vcap/bosh/" + softLayerAgentEnvServiceSettingsFileName
)

type SoftLayerAgentEnvService struct {
	vmId            int
	softLayerClient sl.Client
	logger          boshlog.Logger
}

func NewSoftLayerAgentEnvService(vmId int, softLayerClient sl.Client, logger boshlog.Logger) SoftLayerAgentEnvService {
	return SoftLayerAgentEnvService{
		vmId:            vmId,
		softLayerClient: softLayerClient,
		logger:          logger,
	}
}

func (s SoftLayerAgentEnvService) Fetch() (AgentEnv, error) {
	return AgentEnv{}, errors.New("Implement me!")
}

func (s SoftLayerAgentEnvService) Update(agentEnv AgentEnv) error {
	return errors.New("Implement me!")
}
