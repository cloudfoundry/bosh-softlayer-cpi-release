package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type hasVM struct {
	vmFinder bslcvm.Finder
}

func NewHasVM(vmFinder bslcvm.Finder) Action {
	return &hasVM{vmFinder: vmFinder}
}

func (a *hasVM) Run(vmCID VMCID) (bool, error) {
	_, found, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return false, bosherr.WrapErrorf(err, "Finding VM '%d'", vmCID)
	}

	return found, nil
}
