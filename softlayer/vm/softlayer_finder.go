package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	util "github.com/cloudfoundry/bosh-softlayer-cpi/util"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

const SOFTLAYER_VM_FINDER_LOG_TAG = "SoftLayerVMFinder"

type SoftLayerFinder struct {
	softLayerClient        sl.Client
	agentEnvServiceFactory AgentEnvServiceFactory
	logger                 boshlog.Logger
	uuidGenerator          boshuuid.Generator
	fs                     boshsys.FileSystem
}

func NewSoftLayerFinder(softLayerClient sl.Client, agentEnvServiceFactory AgentEnvServiceFactory, logger boshlog.Logger, uuidGenerator boshuuid.Generator, fs boshsys.FileSystem) SoftLayerFinder {
	return SoftLayerFinder{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		logger:                 logger,
		uuidGenerator:          uuidGenerator,
		fs:                     fs,
	}
}

func (f SoftLayerFinder) Find(vmID int) (VM, bool, error) {
	vm := NewSoftLayerVM(vmID, f.softLayerClient, util.GetSshClient(), f.logger)
	if vm.ID() == 0 {
		return SoftLayerVM{}, false, bosherr.Errorf("Failed to find VM or Baremetal %d", vmID)
	}

	return vm, true, nil
}
