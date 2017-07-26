package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"bosh-softlayer-cpi/registry"
)

type ConcreteFactoryOptions struct {
	Agent    registry.AgentOptions  `json:"agent"`
	Registry registry.ClientOptions `json:"registry,omitempty"`
}

func (o ConcreteFactoryOptions) Validate() error {
	err := o.Agent.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Agent configuration")
	}

	err = o.Registry.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Registry configuration")
	}

	return nil
}
