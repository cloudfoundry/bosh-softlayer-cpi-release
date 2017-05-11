package action

import (
	bslnet "bosh-softlayer-cpi/softlayer/networks"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "bosh-softlayer-cpi/softlayer/common"
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

func (a ConfigureNetworksAction) Run(vmCID VMCID, networks bslnet.Networks) (interface{}, error) {
	_, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding vm '%s'", vmCID)
	}

	//err = vm.ConfigureNetworksSettings(networks)
	//if err != nil {
	//	return nil, bosherr.WrapErrorf(err, "Configuring networks vm '%s'", vmCID)
	//}

	return nil, nil
}
