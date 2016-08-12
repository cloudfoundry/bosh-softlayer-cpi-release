package vm

import (
	"encoding/json"
	"time"

	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const (
	maxAttempts = 30
	delay       = 10
)

type fsAgentEnvService struct {
	vm                   VM
	softlayerFileService SoftlayerFileService
	settingsPath         string
	logger               boshlog.Logger
	logTag               string
}

func NewFSAgentEnvService(
	vm VM,
	softlayerFileService SoftlayerFileService,
	logger boshlog.Logger,
) AgentEnvService {
	return &fsAgentEnvService{
		vm:                   vm,
		softlayerFileService: softlayerFileService,
		settingsPath:         "/var/vcap/bosh/user_data.json",
		logger:               logger,
		logTag:               "FSAgentEnvService",
	}
}

func (s *fsAgentEnvService) Fetch() (AgentEnv, error) {
	var agentEnv AgentEnv

	contents, err := s.softlayerFileService.Download(ROOT_USER_NAME, s.vm.GetRootPassword(), s.vm.GetPrimaryBackendIP(), s.settingsPath)
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

func (s *fsAgentEnvService) Update(agentEnv AgentEnv) error {
	s.logger.Debug(s.logTag, "Updating agent env: %#v", agentEnv)

	jsonBytes, err := json.Marshal(agentEnv)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling agent env")
	}

	for i := 0; i < bslcommon.RETRY_COUNT; i++ {
		s.logger.Debug(s.logTag, "Updating Agent Env: Making attempt #%d", i)
		err = s.softlayerFileService.Upload(ROOT_USER_NAME, s.vm.GetRootPassword(), s.vm.GetPrimaryBackendIP(), s.settingsPath, jsonBytes)
		if err == nil {
			return nil
		}
		time.Sleep(bslcommon.WAIT_TIME)
	}

	// Add this warning message due to bosh-softlayer-cpi issues #129, may remove this piece of code when we identify the real root cause
	var longHostNameWarningMsg = ""
	if bslcommon.LengthOfHostName > 64 {
		longHostNameWarningMsg = "Notice that the length of device hostname is greater than 64 characters, which might cause SSH service setup improperly by SoftLayer, please confirm with SoftLayer or consider to shorten the hostname"
	}
	return bosherr.WrapError(err, "Updating Agent Env timeout. "+longHostNameWarningMsg)
}
