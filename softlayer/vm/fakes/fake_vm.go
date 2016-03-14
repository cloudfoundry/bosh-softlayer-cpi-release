package fakes

import (
	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"

	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type FakeVM struct {
	id int

	DeleteCalled bool
	DeleteErr    error

	RebootCalled bool
	RebootErr    error

	SetMetadataCalled bool
	SetMetadataErr    error
	VMMetadata        bslcvm.VMMetadata

	ConfigureNetworksCalled bool
	ConfigureNetworksErr    error
	Networks                bslcvm.Networks

	AttachDiskDisk bslcdisk.Disk
	AttachDiskErr  error

	DetachDiskDisk bslcdisk.Disk
	DetachDiskErr  error
}

func NewFakeVM(id int) *FakeVM {
	return &FakeVM{id: id}
}

func (vm FakeVM) ID() int { return vm.id }

func (vm *FakeVM) Delete(agentID string) error {
	vm.DeleteCalled = true
	return vm.DeleteErr
}

func (vm *FakeVM) Reboot() error {
	vm.RebootCalled = true
	return vm.RebootErr
}

func (vm *FakeVM) SetMetadata(metadata bslcvm.VMMetadata) error {
	vm.SetMetadataCalled = true
	vm.VMMetadata = metadata
	return vm.SetMetadataErr
}

func (vm *FakeVM) ConfigureNetworks(networks bslcvm.Networks) error {
	vm.ConfigureNetworksCalled = true
	vm.Networks = networks
	return vm.ConfigureNetworksErr
}

func (vm *FakeVM) AttachDisk(disk bslcdisk.Disk) error {
	vm.AttachDiskDisk = disk
	return vm.AttachDiskErr
}

func (vm *FakeVM) DetachDisk(disk bslcdisk.Disk) error {
	vm.DetachDiskDisk = disk
	return vm.DetachDiskErr
}
