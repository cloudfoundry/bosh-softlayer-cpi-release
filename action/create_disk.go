package action

import (
	bosherr "bosh/errors"

	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
)

type CreateDisk struct {
	diskCreator bslcdisk.Creator
}

func NewCreateDisk(diskCreator bslcdisk.Creator) CreateDisk {
	return CreateDisk{diskCreator: diskCreator}
}

func (a CreateDisk) Run(size int, _ VMCID) (DiskCID, error) {
	disk, err := a.diskCreator.Create(size)
	if err != nil {
		return "", bosherr.WrapError(err, "Creating disk of size '%d'", size)
	}

	return DiskCID(disk.ID()), nil
}
