package action

import (
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
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
	_, found, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return false, nil
	}

	return found, nil
}
