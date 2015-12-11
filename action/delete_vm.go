package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type DeleteVM struct {
	vmFinder bslcvm.Finder
}

func NewDeleteVM(vmFinder bslcvm.Finder) DeleteVM {
	return DeleteVM{vmFinder: vmFinder}
}

func (a DeleteVM) Run(vmCID VMCID, agentID string) (interface{}, error) {
	vm, found, err := a.vmFinder.Find(int(vmCID))

	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding vm '%s'", vmCID)
	}

	if found {
		err := vm.Delete(agentID)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Deleting vm '%s'", vmCID)
		}
	}

	return nil, nil
}
