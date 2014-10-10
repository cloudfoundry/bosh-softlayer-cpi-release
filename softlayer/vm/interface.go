package vm

import (
	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

type VMCloudProperties struct {
	StartCpus  int `json:"startCpus,omitempty"`
	MaxMemory  int `json:"maxMemory,omitempty"`
	Datacenter sldatatypes.Datacenter
	SshKeys    []sldatatypes.SshKey `json:"sshKeys"`
}

type Creator interface {
	// Create takes an agent id and creates a VM with provided configuration
	Create(string, bslcstem.Stemcell, VMCloudProperties, Networks, Environment) (VM, error)
}

type Finder interface {
	Find(int) (VM, bool, error)
}

type VM interface {
	ID() int

	Delete() error
	Reboot() error

	AttachDisk(bslcdisk.Disk) error
	DetachDisk(bslcdisk.Disk) error
}

type Environment map[string]interface{}
