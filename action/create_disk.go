package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type createDisk struct {
	diskCreator bslcdisk.Creator
	vmFinder    bslcvm.Finder
}

func NewCreateDisk(vmFinder bslcvm.Finder, diskCreator bslcdisk.Creator) Action {
	return &createDisk{diskCreator: diskCreator, vmFinder: vmFinder}
}

func (a *createDisk) Run(size int, cloudProps bslcdisk.DiskCloudProperties, instanceId VMCID) (string, error) {
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
