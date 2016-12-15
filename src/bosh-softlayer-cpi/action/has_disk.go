package action

import (
	. "bosh-softlayer-cpi/softlayer/disk"
)

type HasDiskAction struct {
	diskFinder DiskFinder
}

func NewHasDisk(
        diskFinder DiskFinder,
) (action HasDiskAction) {
	action.diskFinder = diskFinder
	return
}

func (a HasDiskAction) Run(diskCID DiskCID) (bool, error) {
	_, found, err := a.diskFinder.Find(int(diskCID))
	if err != nil {
		return false, err
	}

	return found, nil
}
