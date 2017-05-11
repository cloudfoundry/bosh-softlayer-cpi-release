package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DeleteVMAction struct {
	vmDeleterProvider DeleterProvider
	options           ConcreteFactoryOptions
}

func NewDeleteVM(
	vmDeleterProvider DeleterProvider,
	options ConcreteFactoryOptions,
) (action DeleteVMAction) {
	action.vmDeleterProvider = vmDeleterProvider
	action.options = options
	return
}

func (a DeleteVMAction) Run(vmCID VMCID) (interface{}, error) {
	vmDeleter := a.vmDeleterProvider.Get("virtualguest")

	err := vmDeleter.Delete(int(vmCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Deleting vm with id %d", int(vmCID))
	}

	return nil, nil
}
