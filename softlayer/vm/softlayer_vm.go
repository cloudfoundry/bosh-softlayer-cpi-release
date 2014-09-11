package vm

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"
	wrdnclient "github.com/cloudfoundry-incubator/garden/client"

	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
)

type SoftLayerVM struct {
	id string

	softLayerClient wrdnclient.Client
	agentEnvService AgentEnvService

	hostBindMounts  HostBindMounts
	guestBindMounts GuestBindMounts

	logger boshlog.Logger
}

func NewSoftLayerVM(
	id string,
	softLayerClient wrdnclient.Client,
	agentEnvService AgentEnvService,
	hostBindMounts HostBindMounts,
	guestBindMounts GuestBindMounts,
	logger boshlog.Logger,
) SoftLayerVM {
	return SoftLayerVM{
		id: id,

		softLayerClient: softLayerClient,
		agentEnvService: agentEnvService,

		hostBindMounts:  hostBindMounts,
		guestBindMounts: guestBindMounts,

		logger: logger,
	}
}

func (vm SoftLayerVM) ID() string { return vm.id }

func (vm SoftLayerVM) Delete() error {
	err := vm.hostBindMounts.DeleteEphemeral(vm.id)
	if err != nil {
		return err
	}

	err = vm.hostBindMounts.DeletePersistent(vm.id)
	if err != nil {
		return err
	}

	return vm.softLayerClient.Destroy(vm.id)
}

func (vm SoftLayerVM) AttachDisk(disk bslcdisk.Disk) error {
	agentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapError(err, "Fetching agent env")
	}

	err = vm.hostBindMounts.MountPersistent(vm.id, disk.ID(), disk.Path())
	if err != nil {
		return bosherr.WrapError(err, "Mounting persistent bind mounts dir")
	}

	diskHintPath := vm.guestBindMounts.MountPersistent(disk.ID())

	agentEnv = agentEnv.AttachPersistentDisk(disk.ID(), diskHintPath)

	err = vm.agentEnvService.Update(agentEnv)
	if err != nil {
		return bosherr.WrapError(err, "Updating agent env")
	}

	return nil
}

func (vm SoftLayerVM) DetachDisk(disk bslcdisk.Disk) error {
	agentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapError(err, "Fetching agent env")
	}

	err = vm.hostBindMounts.UnmountPersistent(vm.id, disk.ID())
	if err != nil {
		return bosherr.WrapError(err, "Unmounting persistent bind mounts dir")
	}

	agentEnv = agentEnv.DetachPersistentDisk(disk.ID())

	err = vm.agentEnvService.Update(agentEnv)
	if err != nil {
		return bosherr.WrapError(err, "Updating agent env")
	}

	return nil
}
