package fakes

import (
	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
	bslcstemcell "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
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

	ReloadOSStemcell bslcstemcell.Stemcell
	ReloadOSErr      error
}

func NewFakeVM(id int) *FakeVM {
	return &FakeVM{id: id}
}

func (vm *FakeVM) ID() int { return vm.id }

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

func (vm *FakeVM) ReloadOS(stemcell bslcstemcell.Stemcell) error {
	vm.ReloadOSStemcell = stemcell
	return vm.ReloadOSErr
}

func (vm *FakeVM) GetDataCenterId() int {
	return 1234
}

func (vm *FakeVM) GetPrimaryIP() string {
	return "127.0.0.1"
}

func (vm *FakeVM) GetPrimaryBackendIP() string {
	return "10.0.0.1"
}

func (vm *FakeVM) GetRootPassword() string {
	return "password"
}

func (vm *FakeVM) SetVcapPassword(encryptedPwd string) error {
	return nil
}
