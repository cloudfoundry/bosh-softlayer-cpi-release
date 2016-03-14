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
	virtualGuestService, err := f.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return SoftLayerVM{}, false, bosherr.WrapError(err, "Creating SoftLayer Virtual Guest Service from client")
	}

	virtualGuest, err := virtualGuestService.GetObject(vmID)
	if err != nil {
		return SoftLayerVM{}, false, bosherr.WrapErrorf(err, "Finding SoftLayer Virtual Guest with id `%d`", vmID)
	}

	vm, found := SoftLayerVM{}, true
	if virtualGuest.Id == vmID {
		softlayerFileService := NewSoftlayerFileService(util.GetSshClient(), virtualGuest, f.logger, f.uuidGenerator, f.fs)
		vm = NewSoftLayerVM(vmID, f.softLayerClient, util.GetSshClient(), f.agentEnvServiceFactory.New(softlayerFileService, strconv.Itoa(vmID)), f.logger)
	} else {
		found = false
	}

	return vm, found, nil
}
