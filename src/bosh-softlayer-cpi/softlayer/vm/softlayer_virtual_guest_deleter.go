package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "bosh-softlayer-cpi/softlayer/common"

	bsl "bosh-softlayer-cpi/softlayer/client"
)

type softLayerVMDeleter struct {
	softLayerClient bsl.Client
	logger          boshlog.Logger
	vmFinder        VMFinder
}

func NewSoftLayerVMDeleter(softLayerClient bsl.Client, logger boshlog.Logger, vmFinder VMFinder) VMDeleter {
	return &softLayerVMDeleter{
		softLayerClient: softLayerClient,
		logger:          logger,
		vmFinder:        vmFinder,
	}
}

func (sd *softLayerVMDeleter) Delete(cid int) error {
	vm, err := sd.vmFinder.Find(cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding VM with id: %d.", cid)
	}

	err = vm.DeleteAgentEnv()
	if err != nil {
		return bosherr.WrapError(err, "Deleting VM's agent env")
	}

	return sd.softLayerClient.CancelInstance(cid)
}
