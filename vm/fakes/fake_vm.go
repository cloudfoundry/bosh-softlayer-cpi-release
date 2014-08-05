package fakes

import (
	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/disk"
)

type FakeVM struct {
	id string

	DeleteCalled bool
	DeleteErr    error

	AttachDiskDisk bslcdisk.Disk
	AttachDiskErr  error

	DetachDiskDisk bslcdisk.Disk
	DetachDiskErr  error
}

func NewFakeVM(id string) *FakeVM {
	return &FakeVM{id: id}
}

func (vm FakeVM) ID() string { return vm.id }

func (vm *FakeVM) Delete() error {
	vm.DeleteCalled = true
	return vm.DeleteErr
}

func (vm *FakeVM) AttachDisk(disk bslcdisk.Disk) error {
	vm.AttachDiskDisk = disk
	return vm.AttachDiskErr
}

func (vm *FakeVM) DetachDisk(disk bslcdisk.Disk) error {
	vm.DetachDiskDisk = disk
	return vm.DetachDiskErr
}
