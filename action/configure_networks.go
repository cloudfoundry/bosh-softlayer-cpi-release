package action

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type ConfigureNetworks struct {
	vmFinder bslcvm.Finder
}

func NewConfigureNetworks(vmFinder bslcvm.Finder) ConfigureNetworks {
	return ConfigureNetworks{
		vmFinder: vmFinder,
	}
}

func (a ConfigureNetworks) Run(vmCID VMCID, networks bslcvm.Networks) (interface{}, error) {
	vm, found, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return nil, bosherr.WrapError(err, "Finding vm '%s'", vmCID)
	}

	if found {
		err := vm.ConfigureNetworks(networks)
		if err != nil {
			return nil, bosherr.WrapError(err, "Configuring networks vm '%s'", vmCID)
		}
	}

	return nil, nil
}
