package action

import (
	. "bosh-softlayer-cpi/softlayer/disk"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
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
	result, err := a.diskFinder.Find(int(diskCID))
	if err != nil {
		return false, bosherr.WrapErrorf(err, "Finding disk with id `%d`", diskCID.Int())
	}

	if result.ID() == 0 {
		return false, bosherr.Errorf("Unable to find disk with id `%d`", diskCID.Int())
	}

	return true, nil
}
