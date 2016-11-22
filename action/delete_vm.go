package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	"fmt"
)

type DeleteVMAction struct {
	vmDeleterProvider DeleterProvider
	options ConcreteFactoryOptions
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
	var err error
	if a.options.Softlayer.FeatureOptions.EnablePool {
		vmDeleter, err = a.vmDeleterProvider.Get("pool")
		if err != nil {
			return nil, bosherr.WrapError(err, "Could not get vm deleter for pool")
		}

		err = vmDeleter.Delete(int(vmCID))
		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Update vm %d to free in pool", int(vmCID)))
		}
	} else {
		vmDeleter, err = a.vmDeleterProvider.Get("virtualguest")
		err = vmDeleter.Delete(int(vmCID))
		if err != nil {
			return nil, bosherr.WrapError(err, "Could not get vm deleter for virtual guest")
		}

		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Deleting vm %d", int(vmCID)))
		}
	}

	return nil, nil
}
