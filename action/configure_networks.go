package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
)

type ConfigureNetworksAction struct {
	vmFinder VMFinder
}

func NewConfigureNetworks(
	vmFinder VMFinder,
) (action ConfigureNetworksAction) {
	action.vmFinder = vmFinder
	return
}

func (a ConfigureNetworksAction) Run(vmCID VMCID, networks Networks) (interface{}, error) {
	vm, found, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding vm '%s'", vmCID)
	}

	if found {
		err := vm.ConfigureNetworks(networks)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Configuring networks vm '%s'", vmCID)
		}
	}

	return nil, nil
}
