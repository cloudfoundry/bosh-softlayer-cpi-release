package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type SetVMMetadata struct {
	vmFinder bslcvm.Finder
}

func NewSetVMMetadata(vmFinder bslcvm.Finder) SetVMMetadata {
	return SetVMMetadata{vmFinder: vmFinder}
}

func (a SetVMMetadata) Run(vmCID VMCID, metadata bslcvm.VMMetadata) (interface{}, error) {
	vm, found, err := a.vmFinder.Find(int(vmCID))
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	if len(metadata) == 0 {
		return nil, nil
	}

	err = vm.SetMetadata(metadata)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Setting metadata '%#v' on VM '%s'", metadata, vmCID)
	}

	return nil, nil
}
