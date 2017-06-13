package action

import (
	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/registry"
	vgs "bosh-softlayer-cpi/softlayer/virtual_guest_service"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boslconfig "bosh-softlayer-cpi/softlayer/config"
)

type DeleteVMAction struct {
	vmService        vgs.Service
	registryClient   registry.Client
	softlayerOptions boslconfig.Config
}

func NewDeleteVM(
	vmDeleterProvider vgs.Service,
	registryClient registry.Client,
	softlayerOptions boslconfig.Config,
) (action DeleteVMAction) {
	action.vmService = vmDeleterProvider
	action.registryClient = registryClient
	action.softlayerOptions = softlayerOptions
	return
}

func (dv DeleteVMAction) Run(vmCID VMCID) (interface{}, error) {
	// Delete the VM
	if err := dv.vmService.Delete(vmCID.Int(), dv.softlayerOptions.EnableVps); err != nil {
		if _, ok := err.(api.CloudError); ok {
			return nil, nil
		}
		return nil, bosherr.WrapErrorf(err, "Deleting vm '%s'", vmCID)
	}

	// Delete the VM agent settings
	if err := dv.registryClient.Delete(vmCID.String()); err != nil {
		return nil, bosherr.WrapErrorf(err, "Deleting vm '%s'", vmCID)
	}

	return nil, nil
}
