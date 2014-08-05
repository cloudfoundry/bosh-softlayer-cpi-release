package vm

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"
	wrdnclient "github.com/cloudfoundry-incubator/garden/client"

	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/disk"
)

type WardenVM struct {
	id string

	wardenClient    wrdnclient.Client
	agentEnvService AgentEnvService

	hostBindMounts  HostBindMounts
	guestBindMounts GuestBindMounts

	logger boshlog.Logger
}

func NewWardenVM(
	id string,
	wardenClient wrdnclient.Client,
	agentEnvService AgentEnvService,
	hostBindMounts HostBindMounts,
	guestBindMounts GuestBindMounts,
	logger boshlog.Logger,
) WardenVM {
	return WardenVM{
		id: id,

		wardenClient:    wardenClient,
		agentEnvService: agentEnvService,

		hostBindMounts:  hostBindMounts,
		guestBindMounts: guestBindMounts,

		logger: logger,
	}
}

func (vm WardenVM) ID() string { return vm.id }

func (vm WardenVM) Delete() error {
	err := vm.hostBindMounts.DeleteEphemeral(vm.id)
	if err != nil {
		return err
	}

	err = vm.hostBindMounts.DeletePersistent(vm.id)
	if err != nil {
		return err
	}

	return vm.wardenClient.Destroy(vm.id)
}

func (vm WardenVM) AttachDisk(disk bslcdisk.Disk) error {
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

func (vm WardenVM) DetachDisk(disk bslcdisk.Disk) error {
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
