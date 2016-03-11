package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type ConcreteFactoryOptions struct {
	Softlayer SoftLayerConfig `json:"softlayer"`

    StemcellsDir  string `json:"stemcelldir,omitempty"`

    Agent bslcvm.AgentOptions `json:"agent"`

	AgentEnvService string `json:"agentenvservice,omitempty`

	Registry        bslcvm.RegistryOptions `json:"registry,omitempty"`
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
	Username string `json:"username"`
	ApiKey   string `json:"apiKey"`
}

func (c SoftLayerConfig) Validate() error {
	if c.Username == "" {
		return bosherr.Error("Must provide non-empty Username")
	}

	if c.ApiKey == "" {
		return bosherr.Error("Must provide non-empty ApiKey")
	}

	return nil
}
