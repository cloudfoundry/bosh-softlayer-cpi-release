package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

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

func (a ConfigureNetworks) Run(vmCID VMCID, networks Networks) (interface{}, error) {
	vm, found, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding vm '%s'", vmCID)
	}

	if found {
		vmNetworks := networks.AsVMNetworks()
		err := vm.ConfigureNetworks(vmNetworks)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Configuring networks vm '%s'", vmCID)
		}
	}

	return nil, nil
}
