package action

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
)

type CreateDisk struct {
	diskCreator bslcdisk.Creator
}

func NewCreateDisk(diskCreator bslcdisk.Creator) CreateDisk {
	return CreateDisk{diskCreator: diskCreator}
}

func (a CreateDisk) Run(size int, virtualGuestId VMCID) (DiskCID, error) {
	disk, err := a.diskCreator.Create(size, int(virtualGuestId))
	if err != nil {
		return 0, bosherr.WrapError(err, "Creating disk of size '%d'", size)
	}

	return DiskCID(disk.ID()), nil
}
