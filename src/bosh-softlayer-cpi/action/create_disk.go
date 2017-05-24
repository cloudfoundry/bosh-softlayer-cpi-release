package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/softlayer/disk_service"
	instance "bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

type CreateDisk struct {
	diskService disk.Service
	vmService   instance.Service
}

func NewCreateDisk(
	diskService disk.Service,
	vmService instance.Service,
) CreateDisk {
	return CreateDisk{
		diskService: diskService,
		vmService:   vmService,
	}
}

func (cd CreateDisk) Run(size int, cloudProps DiskCloudProperties, vmCID VMCID) (string, error) {
	// Find the VM (if provided) so we can create the disk in the same datacenter
	var zone string
	zone = cloudProps.DataCenter
	if vmCID != 0 {
		vm, found, err := cd.vmService.Find(vmCID.Int())
		if err != nil {
			return "", bosherr.WrapError(err, "Creating disk")
		}
		if !found {
			return "", api.NewVMNotFoundError(vmCID.String())
		}

		zone = *vm.Datacenter.Name

	}

	// Create the Disk
	disk, err := cd.diskService.Create(size, cloudProps.Iops, zone)
	if err != nil {
		return "", bosherr.WrapError(err, "Creating disk")
	}

	return DiskCID(disk).String(), nil
}
