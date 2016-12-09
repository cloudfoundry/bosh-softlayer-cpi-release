package common

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	slh "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
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

	SL_CPI_WAIT_TIME_UPDATE_AGENT_ENV, err := strconv.Atoi(os.Getenv("SL_CPI_WAIT_TIME_UPDATE_AGENT_ENV"))
	if err != nil || SL_CPI_WAIT_TIME_UPDATE_AGENT_ENV == 0 {
		SL_CPI_WAIT_TIME_UPDATE_AGENT_ENV = 5
	}
	SL_CPI_RETRY_COUNT_UPDATE_AGENT_ENV, err := strconv.Atoi(os.Getenv("SL_CPI_RETRY_COUNT_UPDATE_AGENT_ENV"))
	if err != nil || SL_CPI_RETRY_COUNT_UPDATE_AGENT_ENV == 0 {
		SL_CPI_RETRY_COUNT_UPDATE_AGENT_ENV = 5
	}

	for i := 0; i < SL_CPI_RETRY_COUNT_UPDATE_AGENT_ENV; i++ {
		s.logger.Debug(s.logTag, "Updating Agent Env: Making attempt #%d", i)
		err = s.softlayerFileService.Upload(ROOT_USER_NAME, s.vm.GetRootPassword(), s.vm.GetPrimaryBackendIP(), s.settingsPath, jsonBytes)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(SL_CPI_WAIT_TIME_UPDATE_AGENT_ENV) * time.Second)
	}

	// Add this warning message due to bosh-softlayer-cpi issues #129, may remove this piece of code when we identify the real root cause
	var longHostNameWarningMsg string
	if slh.LengthOfHostName > 63 {
		longHostNameWarningMsg = "Notice that the length of device hostname is greater than 63 characters, which might cause SSH service setup improperly by SoftLayer, please confirm with SoftLayer or consider to shorten the hostname"
	}

	return bosherr.WrapError(err, "Updating Agent Env timeout. "+longHostNameWarningMsg)
}
