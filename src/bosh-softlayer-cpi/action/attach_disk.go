package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
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

	err = vm.AttachDisk(disk)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Attaching disk '%s' to VM '%s'", diskCID, vmCID)
	}

	return nil, nil
}
