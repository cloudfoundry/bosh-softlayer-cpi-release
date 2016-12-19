package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"fmt"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
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
	var vmDeleter VMDeleter
	if a.options.Softlayer.FeatureOptions.EnablePool {
		vmDeleter = a.vmDeleterProvider.Get("pool")

		err := vmDeleter.Delete(int(vmCID))
		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Update vm %d to free in pool", int(vmCID)))
		}
	} else {
		vmDeleter = a.vmDeleterProvider.Get("virtualguest")

		err := vmDeleter.Delete(int(vmCID))
		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Deleting vm %d", int(vmCID)))
		}
	}

	return nil, nil
}
