package action

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

type CreateVM struct {
	stemcellFinder bslcstem.Finder
	vmCreator      bslcvm.Creator
}

type VMCloudProperties struct {
	StartCpus  int `json:"startCpus,omitempty"`
	MaxMemory  int `json:"maxMemory,omitempty"`
	Datacenter sldatatypes.Datacenter
	SshKeys    []sldatatypes.SshKey `json:"sshKeys"`
}

type Environment map[string]interface{}

func NewCreateVM(stemcellFinder bslcstem.Finder, vmCreator bslcvm.Creator) CreateVM {
	return CreateVM{
		stemcellFinder: stemcellFinder,
		vmCreator:      vmCreator,
	}
}

func (a CreateVM) Run(agentID string, stemcellCID StemcellCID, cloudProps VMCloudProperties, networks Networks, diskIDs []DiskCID, env Environment) (VMCID, error) {
	//DEBUG
	fmt.Println("CreateVM.Run")
	fmt.Printf("----> agentID: %#v\n", agentID)
	fmt.Printf("----> stemcellID: %#v\n", stemcellCID)
	fmt.Printf("----> cloudProps: %#v\n", cloudProps)
	fmt.Printf("----> networks: %#v\n", networks)
	fmt.Printf("----> diskIDs: %#v\n", diskIDs)
	fmt.Printf("----> env: %#v\n", env)
	fmt.Println()
	//DEBUG

	stemcell, found, err := a.stemcellFinder.Find(string(stemcellCID))
	if err != nil {
		return 0, bosherr.WrapError(err, "Finding stemcell '%s'", stemcellCID)
	}

	if !found {
		return 0, bosherr.New("Expected to find stemcell '%s'", stemcellCID)
	}

	vmNetworks := networks.AsVMNetworks()

	vmEnv := bslcvm.Environment(env)

	vm, err := a.vmCreator.Create(agentID, stemcell, vmNetworks, vmEnv)
	if err != nil {
		return 0, bosherr.WrapError(err, "Creating VM with agent ID '%s'", agentID)
	}

	return VMCID(vm.ID()), nil
}
