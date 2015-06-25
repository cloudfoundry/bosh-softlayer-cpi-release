package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type HasVM struct {
	vmFinder bslcvm.Finder
}

func NewHasVM(vmFinder bslcvm.Finder) HasVM {
	return HasVM{vmFinder: vmFinder}
}

func (a HasVM) Run(vmCID VMCID) (bool, error) {
	_, found, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return false, bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	return found, nil
}
