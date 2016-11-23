package action

import (
	"os"
	"strconv"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ConcreteFactoryOptions struct {
	Softlayer SoftLayerConfig `json:"softlayer"`

	Baremetal BaremetalConfig `json:"baremetal,omitempty"`

	Pool PoolConfig `json:"pool,omitempty"`

	StemcellsDir string `json:"stemcelldir,omitempty"`

	Agent AgentOptions `json:"agent"`

	AgentEnvService string `json:"agentenvservice,omitempty"`

	Registry RegistryOptions `json:"registry,omitempty"`
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
	Username       string         `json:"username"`
	ApiKey         string         `json:"apiKey"`
	FeatureOptions FeatureOptions `json:"featureOptions,omitempty"`
}

type BaremetalConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	EndPoint string `json:"endpoint,omitempty"`
}

type PoolConfig struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
}

func (c SoftLayerConfig) Validate() error {
	if c.Username == "" {
		return bosherr.Error("Must provide non-empty Username")
	}

	if c.ApiKey == "" {
		return bosherr.Error("Must provide non-empty ApiKey")
	}

	if c.FeatureOptions.ApiWaitTime == 0 {
		c.FeatureOptions.ApiWaitTime = 0
	}
	err := os.Setenv("SL_API_WAIT_TIME", strconv.Itoa(c.FeatureOptions.ApiWaitTime))
	if err != nil {
		return bosherr.WrapError(err, "Setting Environment Variable")
	}

	if c.FeatureOptions.ApiRetryCount == 0 {
		c.FeatureOptions.ApiRetryCount = 1
	}
	err = os.Setenv("SL_API_RETRY_COUNT", strconv.Itoa(c.FeatureOptions.ApiRetryCount))
	if err != nil {
		return bosherr.WrapError(err, "Setting Environment Variable")
	}

	if c.FeatureOptions.ApiEndpoint == "" {
		c.FeatureOptions.ApiEndpoint = "api.softlayer.com"
	}
	err = os.Setenv("SL_API_ENDPOINT", c.FeatureOptions.ApiEndpoint)
	if err != nil {
		return bosherr.WrapError(err, "Setting Environment Variable")
	}

	if c.FeatureOptions.CreateISCSIVolumeTimeout == 0 {
		c.FeatureOptions.CreateISCSIVolumeTimeout = 600
	}
	err = os.Setenv("SL_CREATE_ISCSI_VOLUME_TIMEOUT", strconv.Itoa(c.FeatureOptions.CreateISCSIVolumeTimeout))
	if err != nil {
		return bosherr.WrapError(err, "Setting Environment Variable")
	}

	if c.FeatureOptions.CreateISCSIVolumePollingInterval == 0 {
		c.FeatureOptions.CreateISCSIVolumePollingInterval = 10
	}
	err = os.Setenv("SL_CREATE_ISCSI_VOLUME_POLLING_INTERVAL", strconv.Itoa(c.FeatureOptions.CreateISCSIVolumePollingInterval))
	if err != nil {
		return bosherr.WrapError(err, "Setting Environment Variable")
	}

	return nil
}
