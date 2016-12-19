package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	slhelper "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"

	bmscl "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	slhw "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/hardware"
	sl "github.com/maximilien/softlayer-go/softlayer"

	"github.com/cloudfoundry/bosh-softlayer-cpi/util"
)

type softLayerFinder struct {
	softLayerClient        sl.Client
	baremetalClient        bmscl.BmpClient
	agentEnvServiceFactory AgentEnvServiceFactory
	logger                 boshlog.Logger
}

func NewSoftLayerFinder(softLayerClient sl.Client, baremetalClient bmscl.BmpClient, agentEnvServiceFactory AgentEnvServiceFactory, logger boshlog.Logger) VMFinder {
	return &softLayerFinder{
		softLayerClient:        softLayerClient,
		baremetalClient:        baremetalClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		logger:                 logger,
	}
}

func (f *softLayerFinder) Find(vmID int) (VM, bool, error) {
	var vm VM
	virtualGuest, err := slhelper.GetObjectDetailsOnVirtualGuest(f.softLayerClient, vmID)
	if err != nil {
		hardware, err := slhelper.GetObjectDetailsOnHardware(f.softLayerClient, vmID)
		if err != nil {
			return nil, false, bosherr.Errorf("Failed to find VM or Baremetal %d", vmID)
		}
		vm = slhw.NewSoftLayerHardware(hardware, f.softLayerClient, f.baremetalClient, util.GetSshClient(), f.logger)
	} else {
		vm = NewSoftLayerVirtualGuest(virtualGuest, f.softLayerClient, util.GetSshClient(), f.logger)
	}

	softlayerFileService := NewSoftlayerFileService(util.GetSshClient(), f.logger)
	agentEnvService := f.agentEnvServiceFactory.New(vm, softlayerFileService)
	vm.SetAgentEnvService(agentEnvService)
	return vm, true, nil
}
