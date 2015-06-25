package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type RebootVM struct {
	vmFinder bslcvm.Finder
}

func NewRebootVM(vmFinder bslcvm.Finder) RebootVM {
	return RebootVM{vmFinder: vmFinder}
}

func (a RebootVM) Run(vmCID VMCID) (interface{}, error) {
	vm, found, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding vm '%s'", vmCID)
	}

	if found {
		err := vm.Reboot()
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Rebooting vm '%s'", vmCID)
		}
	}

	return nil, nil
}
