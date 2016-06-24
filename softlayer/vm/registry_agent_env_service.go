package vm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type registryAgentEnvService struct {
	endpoint string
	logger   boshlog.Logger
	logTag   string
}

type registryResp struct {
	Settings string
}

func NewRegistryAgentEnvService(
	registryOptions RegistryOptions,
	instanceID string,
	logger boshlog.Logger,
) AgentEnvService {
	endpoint := fmt.Sprintf(
		"http://%s:%s@%s:%d/instances/%s/settings",
		registryOptions.Username,
		registryOptions.Password,
		registryOptions.Host,
		registryOptions.Port,
		instanceID,
	)
	return registryAgentEnvService{
		endpoint: endpoint,
		logger:   logger,
		logTag:   "registryAgentEnvService",
	}
}

func (s registryAgentEnvService) Fetch() (AgentEnv, error) {
	s.logger.Debug(s.logTag, "Fetching agent env from registry endpoint %s", s.endpoint)

	httpClient := http.Client{}
	httpResponse, err := httpClient.Get(s.endpoint)
	if err != nil {
		return AgentEnv{}, bosherr.WrapError(err, "Fetching agent env from registry")
	}

	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return AgentEnv{}, bosherr.Errorf("Received non-200 status code when contacting registry: '%d'", httpResponse.StatusCode)
	}

	httpBody, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return AgentEnv{}, bosherr.WrapErrorf(err, "Reading response from registry endpoint '%s'", s.endpoint)
	}

	var resp registryResp

	err = json.Unmarshal(httpBody, &resp)
	if err != nil {
		return AgentEnv{}, bosherr.WrapError(err, "Unmarshalling registry response")
	}

	var agentEnv AgentEnv

	err = json.Unmarshal([]byte(resp.Settings), &agentEnv)
	if err != nil {
		return AgentEnv{}, bosherr.WrapError(err, "Unmarshalling agent env from registry")
	}

	s.logger.Debug(s.logTag, "Received agent env from registry endpoint '%s', contents: '%s'", s.endpoint, httpBody)

	return agentEnv, nil
}

func (s registryAgentEnvService) Update(agentEnv AgentEnv) error {
	settingsJSON, err := json.Marshal(agentEnv)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling agent env")
	}

	s.logger.Debug(s.logTag, "Updating registry endpoint '%s' with agent env: '%s'", s.endpoint, settingsJSON)

	putPayload := bytes.NewReader(settingsJSON)
	request, err := http.NewRequest("PUT", s.endpoint, putPayload)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating PUT request to update registry at '%s' with settings '%s'", s.endpoint, settingsJSON)
	}

	httpClient := http.Client{}
	httpResponse, err := httpClient.Do(request)
	if err != nil {
		return bosherr.WrapErrorf(err, "Updating registry endpoint '%s' with settings: '%s'", s.endpoint, settingsJSON)
	}

	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK && httpResponse.StatusCode != http.StatusCreated {
		return bosherr.Errorf("Received non-2xx status code when contacting registry: '%d'", httpResponse.StatusCode)
	}

	return nil
}
