package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
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
	vm, found, err := a.vmFinder.Find(int(instanceId))
	if err != nil || !found {
		return "0", bosherr.WrapErrorf(err, "Not Finding vm '%s'", instanceId)
	}

	disk, err := a.diskCreator.Create(size, cloudProps, vm.GetDataCenterId())
	if err != nil {
		return "0", bosherr.WrapErrorf(err, "Creating disk of size '%d'", size)
	}

	return DiskCID(disk.ID()).String(), nil
}
