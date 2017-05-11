package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "bosh-softlayer-cpi/softlayer/common"
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
	vm, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding vm with id '%d'", vmCID.Int())
	}

	err = vm.Reboot()
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Rebooting vm with id '%d'", vmCID.Int())
	}

	return nil, nil
}
