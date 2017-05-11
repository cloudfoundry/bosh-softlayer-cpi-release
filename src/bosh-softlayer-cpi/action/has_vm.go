package action

import (
	. "bosh-softlayer-cpi/softlayer/common"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type HasVMAction struct {
	vmFinder VMFinder
}

func NewHasVM(
	vmFinder VMFinder,
) (action HasVMAction) {
	action.vmFinder = vmFinder
	return
}

func (a HasVMAction) Run(vmCID VMCID) (bool, error) {
	vm, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return false, bosherr.WrapErrorf(err, "Finding VM with id `%d`", vmCID.Int())
	}

	if vm.ID() == nil {
		return false, bosherr.Errorf("Unable to find VM with id `%d`", vmCID.Int())
	}

	return true, nil
}
