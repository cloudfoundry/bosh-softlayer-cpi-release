package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
)

type DeleteDisk struct {
	diskFinder bslcdisk.Finder
}

func NewDeleteDisk(diskFinder bslcdisk.Finder) DeleteDisk {
	return DeleteDisk{diskFinder: diskFinder}
}

func (a DeleteDisk) Run(diskCID DiskCID) (interface{}, error) {
	disk, found, err := a.diskFinder.Find(int(diskCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding disk '%s'", diskCID)
	}

	if found {
		err := disk.Delete()
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Deleting disk '%s'", diskCID)
		}
	}

	return nil, nil
}
