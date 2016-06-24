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

	SetVcapPasswordErr error

	ConfigureNetworksCalled bool
	ConfigureNetworksErr    error
	Networks                bslcvm.Networks

	AttachDiskDisk bslcdisk.Disk
	AttachDiskErr  error

	DetachDiskDisk bslcdisk.Disk
	DetachDiskErr  error

	ReloadOSStemcell bslcstemcell.Stemcell
	ReloadOSErr      error

	SetAgentEnvServiceErr error

	UpdateAgentEnvErr error
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

func (vm *FakeVM) ConfigureNetworks2(networks bslcvm.Networks) error {
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
	return 1234567
}

func (vm *FakeVM) GetPrimaryIP() string {
	return "fake-primary-ip"
}

func (vm *FakeVM) GetPrimaryBackendIP() string {
	return "fake-backend-ip"
}

func (vm *FakeVM) GetRootPassword() string {
	return "root-password"
}

func (vm *FakeVM) GetFullyQualifiedDomainName() string {
	return "fake-fullyQualifiedDomainName"
}

func (vm *FakeVM) SetVcapPassword(encryptedPwd string) error {
	return vm.SetVcapPasswordErr
}

func (vm *FakeVM) SetAgentEnvService(agentEnvService bslcvm.AgentEnvService) error {
	return vm.SetAgentEnvServiceErr
}

func (vm *FakeVM) UpdateAgentEnv(agentEnv bslcvm.AgentEnv) error {
	return vm.UpdateAgentEnvErr
}
