package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "bosh-softlayer-cpi/softlayer/common"
	bslcdisk "bosh-softlayer-cpi/softlayer/disk"
)

type CreateDiskAction struct {
	diskCreator bslcdisk.DiskCreator
	vmFinder    VMFinder
}

func NewCreateDisk(
	vmFinder VMFinder,
	diskCreator bslcdisk.DiskCreator,
) (action CreateDiskAction) {
	action.diskCreator = diskCreator
	action.vmFinder = vmFinder
	return
}

func (a CreateDiskAction) Run(size int, cloudProps bslcdisk.DiskCloudProperties, instanceId VMCID) (string, error) {
	vm, err := a.vmFinder.Find(int(instanceId))
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Finding VM with cid '%s'", instanceId)
	}

	disk, err := a.diskCreator.Create(size, cloudProps, *vm.GetDataCenter())
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating disk of size '%d'", size)
	}

	return DiskCID(disk.ID()).String(), nil
}
