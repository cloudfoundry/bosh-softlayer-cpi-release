package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type HasVMAction struct {
	vmFinder bslcvm.Finder
}

func NewHasVM(
	vmFinder bslcvm.Finder,
) (action HasVMAction) {
	action.vmFinder = vmFinder
	return
}

func (a HasVMAction) Run(vmCID VMCID) (bool, error) {
	_, found, err := a.vmFinder.Find(int(vmCID))
	if err != nil || !found {
		return false, bosherr.WrapErrorf(err, "Finding VM '%d'", vmCID)
	}

	return found, nil
}
