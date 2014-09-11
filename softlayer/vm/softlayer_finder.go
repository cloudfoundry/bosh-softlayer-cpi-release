package vm

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"
	wrdnclient "github.com/cloudfoundry-incubator/garden/client"
)

const wardenFinderLogTag = "WardenFinder"

type SoftLayerFinder struct {
	softLayerClient        wrdnclient.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	hostBindMounts  HostBindMounts
	guestBindMounts GuestBindMounts

	logger boshlog.Logger
}

func NewSoftLayerFinder(
	softLayerClient wrdnclient.Client,
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
	f.logger.Debug(wardenFinderLogTag, "Finding container with ID '%s'", id)

	// Cannot just use Lookup(id) since we need to differentiate between error and not found
	containers, err := f.softLayerClient.Containers(nil)
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Listing all containers")
	}

	for _, container := range containers {
		if container.Handle() == id {
			f.logger.Debug(wardenFinderLogTag, "Found container with ID '%s'", id)

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

	f.logger.Debug(wardenFinderLogTag, "Did not find container with ID '%s'", id)

	return nil, false, nil
}
