package action

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type SetVMMetadata struct {
	vmFinder   bslcvm.Finder
	vmMetadata bslcvm.VMMetadata
}

func NewSetVMMetadata(vmFinder bslcvm.Finder) SetVMMetadata {
	return SetVMMetadata{
		vmFinder:   vmFinder,
		vmMetadata: bslcvm.VMMetadata{},
	}
}

func (a SetVMMetadata) Run(vmCID VMCID, metadata bslcvm.VMMetadata) (interface{}, error) {
	vm, found, err := a.vmFinder.Find(int(vmCID))
	if err != nil {
		return nil, bosherr.WrapError(err, "Finding vm '%s'", vmCID)
	}

	if found {
		err := vm.SetMetadata(metadata)
		if err != nil {
			return nil, bosherr.WrapError(err, "Setting metadata on vm '%s'", vmCID)
		}
	}

	return nil, nil
}
