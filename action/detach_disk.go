package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type DetachDiskAction struct {
	vmFinder   bslcvm.Finder
	diskFinder bslcdisk.DiskFinder
}

func NewDetachDisk(
	vmFinder bslcvm.Finder,
	diskFinder bslcdisk.DiskFinder,
) (action DetachDiskAction) {
	action.vmFinder = vmFinder
	action.diskFinder = diskFinder
	return
}

func (a DetachDiskAction) Run(vmCID VMCID, diskCID DiskCID) (interface{}, error) {
	vm, found, err := a.vmFinder.Find(vmCID.Int())
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	if !found {
		return nil, bosherr.Errorf("Expected to find VM '%s'", vmCID)
	}

	disk, found, err := a.diskFinder.Find(diskCID.Int())
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding disk '%s'", diskCID)
	}

	if !found {
		return nil, bosherr.Errorf("Expected to find disk '%s'", diskCID)
	}

	err = vm.DetachDisk(disk)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Detaching disk '%s' from VM '%s'", diskCID, vmCID)
	}

	return nil, nil
}
