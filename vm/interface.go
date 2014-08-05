package vm

import (
	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/disk"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/stemcell"
)

type Creator interface {
	// Create takes an agent id and creates a VM with provided configuration
	Create(string, bslcstem.Stemcell, Networks, Environment) (VM, error)
}

type Finder interface {
	Find(string) (VM, bool, error)
}

type VM interface {
	ID() string

	Delete() error

	AttachDisk(bslcdisk.Disk) error
	DetachDisk(bslcdisk.Disk) error
}

type Environment map[string]interface{}
