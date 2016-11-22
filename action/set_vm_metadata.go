package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
)

type SetVMMetadataAction struct {
	vmFinder VMFinder
}

func NewSetVMMetadata(
	vmFinder VMFinder,
) (action SetVMMetadataAction) {
	action.vmFinder = vmFinder
	return
}

func (a SetVMMetadataAction) Run(vmCID VMCID, metadata VMMetadata) (interface{}, error) {
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
