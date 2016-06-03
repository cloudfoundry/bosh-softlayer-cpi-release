package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type deleteVM struct {
	vmFinder bslcvm.Finder
}

func NewDeleteVM(vmFinder bslcvm.Finder) Action {
	return &deleteVM{vmFinder: vmFinder}
}

func (a *deleteVM) Run(vmCID VMCID) (interface{}, error) {
	vm, found, err := a.vmFinder.Find(int(vmCID))

	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding vm '%s'", vmCID)
	}

	if found {
		err := vm.Delete("")
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Deleting vm '%s'", vmCID)
		}
	}

	return nil, nil
}
