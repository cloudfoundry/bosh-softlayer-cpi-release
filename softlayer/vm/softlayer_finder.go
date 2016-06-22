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
	var vm VM
	virtualGuest, err := bslcommon.GetObjectDetailsOnVirtualGuest(f.softLayerClient, vmID)
	if err != nil {
		hardware, err := bslcommon.GetObjectDetailsOnHardware(f.softLayerClient, vmID)
		if err != nil {
			return nil, false, bosherr.Errorf("Failed to find VM or Baremetal %d", vmID)
		}
		vm = NewSoftLayerHardware(hardware, f.softLayerClient, f.baremetalClient, util.GetSshClient(), f.logger)
	} else {
		vm = NewSoftLayerVirtualGuest(virtualGuest, f.softLayerClient, util.GetSshClient(), f.logger)
	}

	softlayerFileService := NewSoftlayerFileService(util.GetSshClient(), f.logger)
	agentEnvService := f.agentEnvServiceFactory.New(vm, softlayerFileService)
	vm.SetAgentEnvService(agentEnvService)
	return vm, true, nil
}
