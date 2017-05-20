package action

import (
	"bosh-softlayer-cpi/api"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	instance "bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

type RebootVM struct {
	vmService instance.Service
}

func NewRebootVM(
	vmService instance.Service,
) RebootVM {
	return RebootVM{
		vmService: vmService,
	}
}

func (rv RebootVM) Run(vmCID VMCID) (interface{}, error) {
	if err := rv.vmService.Reboot(vmCID.Int()); err != nil {
		if _, ok := err.(api.CloudError); ok {
			return nil, err
		}
		return nil, bosherr.WrapErrorf(err, "Rebooting vm '%s'", vmCID)
	}

	return nil, nil
}
