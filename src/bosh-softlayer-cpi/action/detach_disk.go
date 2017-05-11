package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "bosh-softlayer-cpi/softlayer/common"
	bslcdisk "bosh-softlayer-cpi/softlayer/disk"
)

type DetachDiskAction struct {
	vmFinder   VMFinder
	diskFinder bslcdisk.DiskFinder
}

func NewDetachDisk(
	vmFinder VMFinder,
	diskFinder bslcdisk.DiskFinder,
) (action DetachDiskAction) {
	action.vmFinder = vmFinder
	action.diskFinder = diskFinder
	return
}

func (a DetachDiskAction) Run(vmCID VMCID, diskCID DiskCID) (interface{}, error) {
	vm, err := a.vmFinder.Find(vmCID.Int())
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	disk, err := a.diskFinder.Find(diskCID.Int())
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding disk '%s'", diskCID)
	}

	err = vm.DetachDisk(disk)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Detaching disk '%s' from VM '%s'", diskCID, vmCID)
	}

	return nil, nil
}
