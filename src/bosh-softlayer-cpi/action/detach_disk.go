package action

import (
	bosherr "github.com/bluebosh/bosh-utils/errors"

	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/softlayer/virtual_guest_service"

	"bosh-softlayer-cpi/registry"
)

type DetachDisk struct {
	vmService      instance.Service
	registryClient registry.Client
}

func NewDetachDisk(
	vmService instance.Service,
	registryClient registry.Client,
) DetachDisk {
	return DetachDisk{
		vmService:      vmService,
		registryClient: registryClient,
	}
}

func (dd DetachDisk) Run(vmCID VMCID, diskCID DiskCID) (interface{}, error) {
	// Detach the disk
	if err := dd.vmService.DetachDisk(vmCID.Int(), diskCID.Int()); err != nil {
		if _, ok := err.(api.CloudError); ok {
			return nil, err
		}
		return nil, bosherr.WrapErrorf(err, "Detaching disk '%s' from vm '%s", diskCID, vmCID)
	}

	exist, err := dd.registryClient.IsExist(vmCID.String())
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Checking if instance '%s' exists", vmCID)
	}
	if !exist {
		return nil, nil
	}

	// Read VM agent settings
	oldAgentSettings, err := dd.registryClient.Fetch(vmCID.String())
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Detaching disk '%s' from vm '%s'", diskCID, vmCID)
	}

	// Update VM agent settings
	newAgentSettings := oldAgentSettings.DetachPersistentDisk(diskCID.String())
	if err = dd.registryClient.Update(vmCID.String(), newAgentSettings); err != nil {
		return nil, bosherr.WrapErrorf(err, "Detaching disk '%s' from vm '%s'", diskCID, vmCID)
	}

	return nil, nil
}
