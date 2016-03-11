package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	util "github.com/cloudfoundry/bosh-softlayer-cpi/util"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

const softLayerFinderLogTag = "SoftLayerFinder"

type SoftLayerFinder struct {
	softLayerClient        sl.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	logger boshlog.Logger
}

func NewSoftLayerFinder(softLayerClient sl.Client, agentEnvServiceFactory AgentEnvServiceFactory, logger boshlog.Logger) SoftLayerFinder {
	return SoftLayerFinder{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,

		logger: logger,
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
		vm = NewSoftLayerVM(vmID, f.softLayerClient, util.GetSshClient(), f.agentEnvServiceFactory.New(vmID), f.logger)
	} else {
		found = false
	}

	return vm, found, nil
}
