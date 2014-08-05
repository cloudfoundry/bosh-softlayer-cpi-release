package action

import (
	bosherr "bosh/errors"

	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type CreateVM struct {
	stemcellFinder bslcstem.Finder
	vmCreator      bslcvm.Creator
}

type ResourcePool struct{}

type Environment map[string]interface{}

func NewCreateVM(stemcellFinder bslcstem.Finder, vmCreator bslcvm.Creator) CreateVM {
	return CreateVM{
		stemcellFinder: stemcellFinder,
		vmCreator:      vmCreator,
	}
}

func (a CreateVM) Run(agentID string, stemcellCID StemcellCID, _ ResourcePool, networks Networks, _ []DiskCID, env Environment) (VMCID, error) {
	stemcell, found, err := a.stemcellFinder.Find(string(stemcellCID))
	if err != nil {
		return "", bosherr.WrapError(err, "Finding stemcell '%s'", stemcellCID)
	}

	if !found {
		return "", bosherr.New("Expected to find stemcell '%s'", stemcellCID)
	}

	vmNetworks := networks.AsVMNetworks()

	vmEnv := bslcvm.Environment(env)

	vm, err := a.vmCreator.Create(agentID, stemcell, vmNetworks, vmEnv)
	if err != nil {
		return "", bosherr.WrapError(err, "Creating VM with agent ID '%s'", agentID)
	}

	return VMCID(vm.ID()), nil
}
