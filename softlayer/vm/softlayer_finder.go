package vm

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

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
	accountService, err := f.softLayerClient.GetSoftLayer_Account_Service()
	if err != nil {
		return SoftLayerVM{}, false, bosherr.WrapError(err, "Creating SoftLayer AcccountService from client")
	}

	virtualGuests, err := accountService.GetVirtualGuests()
	if err != nil {
		return SoftLayerVM{}, false, bosherr.WrapError(err, "Getting a list of SoftLayer VirtualGuests from client")
	}

	found, vm := false, SoftLayerVM{}
	for _, virtualGuest := range virtualGuests {
		if virtualGuest.Id == vmID {
			vm = NewSoftLayerVM(vmID, f.softLayerClient, f.agentEnvServiceFactory.New(vmID), f.logger)
			found = true
			break
		}
	}

	return vm, found, nil
}
