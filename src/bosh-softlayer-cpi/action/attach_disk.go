package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "bosh-softlayer-cpi/softlayer/common"
	bslcdisk "bosh-softlayer-cpi/softlayer/disk"
)

type AttachDiskAction struct {
	vmFinder   VMFinder
	diskFinder bslcdisk.DiskFinder
}

func NewAttachDisk(
	vmFinder VMFinder,
	diskFinder bslcdisk.DiskFinder,
) (action AttachDiskAction) {
	action.vmFinder = vmFinder
	action.diskFinder = diskFinder
	return
}

func (a AttachDiskAction) Run(vmCID VMCID, diskCID DiskCID) (interface{}, error) {
	vm, err := a.vmFinder.Find(vmCID.Int())
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	disk, err := a.diskFinder.Find(diskCID.Int())
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding disk '%s'", diskCID)
	}

	err = vm.AttachDisk(disk)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Attaching disk '%s' to VM '%s'", diskCID, vmCID)
	}

	return nil, nil
}
