package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type ConcreteFactoryOptions struct {
	StemcellsDir string

	Agent bslcvm.AgentOptions

	AgentEnvService string
	Registry        bslcvm.RegistryOptions
}

func (o ConcreteFactoryOptions) Validate() error {
	if o.StemcellsDir == "" {
		return bosherr.Error("Must provide non-empty StemcellsDir")
	}

	err := o.Agent.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Agent configuration")
	}

	return nil
}
