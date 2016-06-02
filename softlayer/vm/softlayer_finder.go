package vm

import (
	"strconv"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	util "github.com/cloudfoundry/bosh-softlayer-cpi/util"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

const SOFTLAYER_VM_FINDER_LOG_TAG = "SoftLayerVMFinder"

type SoftLayerFinder struct {
	softLayerClient        sl.Client
	agentEnvServiceFactory AgentEnvServiceFactory
	logger                 boshlog.Logger
}

func NewSoftLayerFinder(softLayerClient sl.Client, agentEnvServiceFactory AgentEnvServiceFactory, logger boshlog.Logger) SoftLayerFinder {
	return SoftLayerFinder{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		logger:                 logger,
	}
}

func (f SoftLayerFinder) Find(vmID int) (VM, bool, error) {
	virtualGuestService, err := f.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return SoftLayerVM{}, false, bosherr.WrapError(err, "Creating SoftLayer Virtual Guest Service from client")
	}

	virtualGuest, err := virtualGuestService.GetObject(vmID)
	if err != nil {
		if !strings.Contains(err.Error(), "HTTP error code") {
			return SoftLayerVM{}, false, bosherr.WrapErrorf(err, "Finding SoftLayer Virtual Guest with id `%d`", vmID)
		}
	}

	vm, found := SoftLayerVM{}, true
	if virtualGuest.Id == vmID {
		softlayerFileService := NewSoftlayerFileService(util.GetSshClient(), virtualGuest, f.logger)
		vm = NewSoftLayerVM(vmID, f.softLayerClient, util.GetSshClient(), f.agentEnvServiceFactory.New(softlayerFileService, strconv.Itoa(vmID)), f.logger)
	} else {
		found = false
	}

	return vm, found, nil
}
