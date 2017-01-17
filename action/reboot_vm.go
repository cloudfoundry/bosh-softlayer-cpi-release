package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
)

type RebootVMAction struct {
	vmFinder VMFinder
}

func NewRebootVM(
	vmFinder VMFinder,
) (action RebootVMAction) {
	action.vmFinder = vmFinder
	return
}

func (a RebootVMAction) Run(vmCID VMCID) (interface{}, error) {
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
