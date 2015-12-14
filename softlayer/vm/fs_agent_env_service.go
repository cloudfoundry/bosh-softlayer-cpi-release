package vm

import (
	"encoding/json"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const (
	maxAttempts = 30
	delay       = 10
)

type fsAgentEnvService struct {
	softlayerFileService SoftlayerFileService
	settingsPath         string
	logger               boshlog.Logger
	logTag               string
}

func NewFSAgentEnvService(
	softlayerFileService SoftlayerFileService,
	logger boshlog.Logger,
) AgentEnvService {
	return fsAgentEnvService{
		softlayerFileService: softlayerFileService,
		settingsPath:         "/var/vcap/bosh/user_data.json",
		logger:               logger,
		logTag:               "FSAgentEnvService",
	}
}

func (s fsAgentEnvService) Fetch() (AgentEnv, error) {
	var agentEnv AgentEnv

	contents, err := s.softlayerFileService.Download(s.settingsPath)
	if err != nil {
		return AgentEnv{}, bosherr.WrapError(err, "Downloading agent env from virtual guestr")
	}

	err = json.Unmarshal(contents, &agentEnv)
	if err != nil {
		return AgentEnv{}, bosherr.WrapError(err, "Unmarshalling agent env")
	}

	s.logger.Debug(s.logTag, "Fetched agent env: %#v", agentEnv)

	return agentEnv, nil
}

func (s fsAgentEnvService) Update(agentEnv AgentEnv) error {
	s.logger.Debug(s.logTag, "Updating agent env: %#v", agentEnv)

	jsonBytes, err := json.Marshal(agentEnv)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling agent env")
	}

	for i := 0; i < maxAttempts; i++ {
		s.logger.Debug(s.logTag, "Updating Agent Env: Making attempt #%d", i)
		err = s.softlayerFileService.Upload(s.settingsPath, jsonBytes)
		if err == nil {
			return nil
		}
		time.Sleep(delay * time.Second)
	}
	return bosherr.WrapError(err, "Updating Agent Env timeout")
}
