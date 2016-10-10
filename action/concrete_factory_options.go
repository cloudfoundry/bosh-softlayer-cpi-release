package action

import (
	"os"

	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ConcreteFactoryOptions struct {
	Softlayer SoftLayerConfig `json:"softlayer"`

	Baremetal BaremetalConfig `json:"baremetal,omitempty"`

	StemcellsDir string `json:"stemcelldir,omitempty"`

	Agent bslcvm.AgentOptions `json:"agent"`

	AgentEnvService string `json:"agentenvservice,omitempty"`

	Registry bslcvm.RegistryOptions `json:"registry,omitempty"`
}

func (o ConcreteFactoryOptions) Validate() error {
	err := o.Agent.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Agent configuration")
	}

	err = o.Softlayer.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating SoftLayer configuration")
	}

	return nil
}

type SoftLayerConfig struct {
	Username                         string `json:"username"`
	ApiKey                           string `json:"apiKey"`
	ApiEndpoint                      string `json:"apiEndpoint,omitempty"`
	ApiWaitTime                      string `json:"apiWaitTime,omitempty"`
	ApiRetryCount                    string `json:"apiRetryCount,omitempty"`
	CreateISCSIVolumeTimeout         string `json:"createIscsiVolumeTimeout,omitempty"`
	CreateISCSIVolumePollingIntreval string `json:"createIscsiVolumePollingIntertval,omitempty"`
}

type BaremetalConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	EndPoint string `json:"endpoint,omitempty"`
}

func (c SoftLayerConfig) Validate() error {
	if c.Username == "" {
		return bosherr.Error("Must provide non-empty Username")
	}

	if c.ApiKey == "" {
		return bosherr.Error("Must provide non-empty ApiKey")
	}

	err := os.Setenv("SL_API_WAIT_TIME", c.ApiWaitTime)
	if err != nil {
		return bosherr.WrapError(err, "Setting Environment Variable")
	}

	err = os.Setenv("SL_API_RETRY_COUNT", c.ApiRetryCount)
	if err != nil {
		return bosherr.WrapError(err, "Setting Environment Variable")
	}

	err = os.Setenv("SL_API_ENDPOINT", c.ApiEndpoint)
	if err != nil {
		return bosherr.WrapError(err, "Setting Environment Variable")
	}

	err = os.Setenv("SL_CREATE_ISCSI_VOLUME_TIMEOUT", c.CreateISCSIVolumeTimeout)
	if err != nil {
		return bosherr.WrapError(err, "Setting Environment Variable")
	}

	err = os.Setenv("SL_CREATE_ISCSI_VOLUME_POLLING_INTERVAL", c.CreateISCSIVolumePollingIntreval)
	if err != nil {
		return bosherr.WrapError(err, "Setting Environment Variable")
	}

	return nil
}
