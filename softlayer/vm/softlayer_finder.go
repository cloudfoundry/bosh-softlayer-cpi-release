package vm

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bslcpi "github.com/maximilien/bosh-softlayer-cpi/softlayer/cpi"
)

const softLayerFinderLogTag = "SoftLayerFinder"

type SoftLayerFinder struct {
	softLayerClient        bslcpi.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	hostBindMounts  HostBindMounts
	guestBindMounts GuestBindMounts

	logger boshlog.Logger
}

func NewSoftLayerFinder(
	softLayerClient bslcpi.Client,
	agentEnvServiceFactory AgentEnvServiceFactory,
	hostBindMounts HostBindMounts,
	guestBindMounts GuestBindMounts,
	logger boshlog.Logger,
) SoftLayerFinder {
	return SoftLayerFinder{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,

		hostBindMounts:  hostBindMounts,
		guestBindMounts: guestBindMounts,

		logger: logger,
	}
}

func (f SoftLayerFinder) Find(id string) (VM, bool, error) {
	f.logger.Debug(softLayerFinderLogTag, "Finding container with ID '%s'", id)

	// Cannot just use Lookup(id) since we need to differentiate between error and not found
	containers, err := f.softLayerClient.Containers()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Listing all containers")
	}

	for _, container := range containers {
		if container.Handle() == id {
			f.logger.Debug(softLayerFinderLogTag, "Found container with ID '%s'", id)

			agentEnvService := f.agentEnvServiceFactory.New(container)

			vm := NewSoftLayerVM(
				id,
				f.softLayerClient,
				agentEnvService,
				f.hostBindMounts,
				f.guestBindMounts,
				f.logger,
			)

			return vm, true, nil
		}
	}

	f.logger.Debug(softLayerFinderLogTag, "Did not find container with ID '%s'", id)

	return nil, false, nil
}
