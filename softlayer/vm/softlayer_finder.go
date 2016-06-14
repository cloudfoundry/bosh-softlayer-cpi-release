package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"

	bmscl "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	"github.com/cloudfoundry/bosh-softlayer-cpi/util"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

type softLayerFinder struct {
	softLayerClient        sl.Client
	baremetalClient        bmscl.BmpClient
	agentEnvServiceFactory AgentEnvServiceFactory
	logger                 boshlog.Logger
}

func NewSoftLayerFinder(softLayerClient sl.Client, baremetalClient bmscl.BmpClient, agentEnvServiceFactory AgentEnvServiceFactory, logger boshlog.Logger) Finder {
	return &softLayerFinder{
		softLayerClient:        softLayerClient,
		baremetalClient:        baremetalClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		logger:                 logger,
	}
}

func (f *softLayerFinder) Find(vmID int) (VM, bool, error) {
	softlayerFileService := NewSoftlayerFileService(util.GetSshClient(), f.logger)
	agentEnvService := f.agentEnvServiceFactory.New(softlayerFileService)

	_, err := bslcommon.GetObjectDetailsOnVirtualGuest(f.softLayerClient, vmID)
	if err != nil {
		_, err := bslcommon.GetObjectDetailsOnHardware(f.softLayerClient, vmID)
		if err != nil {
			return nil, false, bosherr.Errorf("Failed to find VM or Baremetal %d", vmID)
		}
		vm := NewSoftLayerHardware(vmID, f.softLayerClient, f.baremetalClient, util.GetSshClient(), agentEnvService, f.logger)
		softlayerFileService.SetVM(vm)
		return vm, true, nil
	}

	vm := NewSoftLayerVirtualGuest(vmID, f.softLayerClient, util.GetSshClient(), agentEnvService, f.logger)
	softlayerFileService.SetVM(vm)
	return vm, true, nil
}
