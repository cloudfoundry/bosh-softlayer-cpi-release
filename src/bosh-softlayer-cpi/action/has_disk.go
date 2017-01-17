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
	result, found, err := a.diskFinder.Find(int(diskCID))
	if err != nil {
		return false, err
	} else {
		if result == nil {
			return false, nil
		}
	}

	return found, nil
}
