package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/softlayer/disk_service"
	"bosh-softlayer-cpi/softlayer/virtual_guest_service"

	"bosh-softlayer-cpi/registry"
)

type AttachDisk struct {
	diskService    disk.Service
	vmService      instance.Service
	registryClient registry.Client
}

func NewAttachDisk(
	diskService disk.Service,
	vmService instance.Service,
	registryClient registry.Client,
) AttachDisk {
	return AttachDisk{
		diskService:    diskService,
		vmService:      vmService,
		registryClient: registryClient,
	}
}

func (ad AttachDisk) Run(vmCID VMCID, diskCID DiskCID) (interface{}, error) {
	// Find the disk
	_, err := ad.diskService.Find(diskCID.Int())
	if err != nil {
		if _, ok := err.(api.CloudError); ok {
			return nil, err
		}
		return nil, bosherr.WrapErrorf(err, "Attaching disk '%s' to vm '%s", diskCID, vmCID)
	}

	// Attach the Disk to the VM
	persistentSetting, err := ad.vmService.AttachDisk(vmCID.Int(), diskCID.Int())
	if err != nil {
		if _, ok := err.(api.CloudError); ok {
			return nil, err
		}
		return nil, bosherr.WrapErrorf(err, "Attaching disk '%s' to vm '%s", diskCID, vmCID)
	}

	// Read VM agent settings
	agentSettings, err := ad.registryClient.Fetch(vmCID.String())
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Attaching disk '%s' to vm '%s'", diskCID, vmCID)
	}

	// Update VM agent settings
	newAgentSettings, err := agentSettings.AttachPersistentDisk(diskCID.String(), persistentSetting)
	if err = ad.registryClient.Update(vmCID.String(), newAgentSettings); err != nil {
		return nil, bosherr.WrapErrorf(err, "Attaching disk '%s' to vm '%s'", diskCID, vmCID)
	}

	return nil, nil
}
