package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
)

type CreateDisk struct {
	diskCreator bslcdisk.Creator
}

func NewCreateDisk(diskCreator bslcdisk.Creator) CreateDisk {
	return CreateDisk{diskCreator: diskCreator}
}

func (a CreateDisk) Run(size int, cloudProps bslcdisk.DiskCloudProperties, instanceId VMCID) (string, error) {
	disk, err := a.diskCreator.Create(size, cloudProps, instanceId.Int())
	if err != nil {
		return "0", bosherr.WrapErrorf(err, "Creating disk of size '%d'", size)
	}

	return DiskCID(disk.ID()).String(), nil
}
