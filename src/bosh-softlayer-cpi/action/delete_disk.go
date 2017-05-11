package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcdisk "bosh-softlayer-cpi/softlayer/disk"
)

type DeleteDiskAction struct {
	diskFinder bslcdisk.DiskFinder
}

func NewDeleteDisk(
	diskFinder bslcdisk.DiskFinder,
) (action DeleteDiskAction) {
	action.diskFinder = diskFinder
	return
}

func (a DeleteDiskAction) Run(diskCID DiskCID) (interface{}, error) {
	disk, err := a.diskFinder.Find(int(diskCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding disk '%s'", diskCID)
	}

	err = disk.Delete()
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Deleting disk '%s'", diskCID)
	}

	return nil, nil
}
