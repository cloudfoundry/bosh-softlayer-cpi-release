package action

import (
	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/registry"
	vgs "bosh-softlayer-cpi/softlayer/virtual_guest_service"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DeleteVMAction struct {
	vmService      vgs.SoftlayerVirtualGuestService
	registryClient registry.Client
}

func NewDeleteVM(
	vmDeleterProvider vgs.SoftlayerVirtualGuestService,
	registryClient registry.Client,
) (action DeleteVMAction) {
	action.vmService = vmDeleterProvider
	action.registryClient = registryClient
	return
}

func (dv DeleteVMAction) Run(vmCID VMCID) (interface{}, error) {
	// Delete the VM
	if err := dv.vmService.Delete(vmCID.Int()); err != nil {
		if _, ok := err.(api.CloudError); ok {
			return nil, err
		}
		return nil, bosherr.WrapErrorf(err, "Deleting vm '%s'", vmCID)
	}

	// Delete the VM agent settings
	if err := dv.registryClient.Delete(vmCID.String()); err != nil {
		return nil, bosherr.WrapErrorf(err, "Deleting vm '%s'", vmCID)
	}

	return nil, nil
}
