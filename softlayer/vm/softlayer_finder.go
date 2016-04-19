package vm

import (
        "strconv"   
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
	softlayerFileService := NewSoftlayerFileService1(util.GetSshClient(), f.logger, f.uuidGenerator, f.fs)
	agentEnvService := f.agentEnvServiceFactory.New(softlayerFileService, strconv.Itoa(vmID))
	vm := NewSoftLayerVM(vmID, f.softLayerClient, util.GetSshClient(), agentEnvService, f.logger)
f.logger.Debug(SOFTLAYER_VM_FINDER_LOG_TAG, "Object Jimmy attachdisk 44 %v ", vm)
	if vm.ID() == 0 {
		return SoftLayerVM{}, false, bosherr.Errorf("Failed to find VM or Baremetal %d", vmID)
	}
        softlayerFileService.SetVM( vm )
f.logger.Debug(SOFTLAYER_VM_FINDER_LOG_TAG, "Object Jimmy attachdisk 55 %v ", vm)

	return vm, true, nil
}
