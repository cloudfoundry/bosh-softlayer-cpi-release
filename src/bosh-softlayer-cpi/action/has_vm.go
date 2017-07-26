package action

import (
	instance "bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

type HasVM struct {
	vmService instance.Service
}

func NewHasVM(
	vmService instance.Service,
) HasVM {
	return HasVM{
		vmService: vmService,
	}
}

func (hv HasVM) Run(vmCID VMCID) (bool, error) {
	_, err := hv.vmService.Find(vmCID.Int())
	if err != nil {
		return false, err
	}

	return true, nil
}
