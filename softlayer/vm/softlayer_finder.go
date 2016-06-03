package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"strconv"

	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"

	"github.com/cloudfoundry/bosh-softlayer-cpi/util"
	sl "github.com/maximilien/softlayer-go/softlayer"
	bmscl "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
)

type softLayerFinder struct {
	softLayerClient        sl.Client
	baremetalClient        bmscl.BmpClient
	agentEnvServiceFactory AgentEnvServiceFactory
	logger                 boshlog.Logger
	uuidGenerator          boshuuid.Generator
	fs                     boshsys.FileSystem
}

func NewSoftLayerFinder(softLayerClient sl.Client, baremetalClient bmscl.BmpClient, agentEnvServiceFactory AgentEnvServiceFactory, logger boshlog.Logger, uuidGenerator boshuuid.Generator, fs boshsys.FileSystem) Finder {
	return &softLayerFinder{
		softLayerClient:        softLayerClient,
		baremetalClient:        baremetalClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		logger:                 logger,
		uuidGenerator:          uuidGenerator,
		fs:                     fs,
	}
}

func (f *softLayerFinder) Find(vmID int) (VM, bool, error) {
	softlayerFileService := NewSoftlayerFileService(util.GetSshClient(), f.logger, f.uuidGenerator, f.fs)
	agentEnvService := f.agentEnvServiceFactory.New(softlayerFileService, strconv.Itoa(vmID))

	_, err := bslcommon.GetObjectDetailsOnVirtualGuest(f.softLayerClient, vmID)
	if err != nil {
		_, err := bslcommon.GetObjectDetailsOnHardware(f.softLayerClient, vmID)
		if err != nil {
			return nil, false, bosherr.Errorf("Failed to find VM or Baremetal %d", vmID)
		}
		vm := NewSoftLayerHardware(vmID, f.softLayerClient,f.baremetalClient, util.GetSshClient(), agentEnvService, f.logger)
		softlayerFileService.SetVM(vm)
		return vm, true, nil
	}

	vm := NewSoftLayerVirtualGuest(vmID, f.softLayerClient, util.GetSshClient(), agentEnvService, f.logger)
	softlayerFileService.SetVM(vm)
	return vm, true, nil
}
