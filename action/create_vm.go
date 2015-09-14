package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type CreateVM struct {
	stemcellFinder    bslcstem.Finder
	vmCreator         bslcvm.Creator
	vmCloudProperties bslcvm.VMCloudProperties
}

type Environment map[string]interface{}

func NewCreateVM(stemcellFinder bslcstem.Finder, vmCreator bslcvm.Creator) CreateVM {
	return CreateVM{
		stemcellFinder:    stemcellFinder,
		vmCreator:         vmCreator,
		vmCloudProperties: bslcvm.VMCloudProperties{},
	}
}

func (a CreateVM) Run(agentID string, stemcellCID StemcellCID, cloudProps bslcvm.VMCloudProperties, networks Networks, diskIDs []DiskCID, env Environment) (string, error) {
	a.updateCloudProperties(cloudProps)

	stemcell, found, err := a.stemcellFinder.FindById(int(stemcellCID))
	if err != nil {
		return "0", bosherr.WrapErrorf(err, "Finding stemcell '%s'", stemcellCID)
	}

	if !found {
		return "0", bosherr.Errorf("Expected to find stemcell '%s'", stemcellCID)
	}

	vmNetworks := networks.AsVMNetworks()

	vmEnv := bslcvm.Environment(env)

	vm, err := a.vmCreator.Create(agentID, stemcell, cloudProps, vmNetworks, vmEnv)
	if err != nil {
		return "0", bosherr.WrapErrorf(err, "Creating VM with agent ID '%s'", agentID)
	}

	return VMCID(vm.ID()).String(), nil
}

func (a CreateVM) updateCloudProperties(cloudProps bslcvm.VMCloudProperties) {
	if cloudProps.StartCpus > 1 {
		a.vmCloudProperties.StartCpus = cloudProps.StartCpus
	}

	if cloudProps.MaxMemory > 1024 {
		a.vmCloudProperties.MaxMemory = cloudProps.MaxMemory
	}

	if cloudProps.Datacenter.Name != a.vmCloudProperties.Datacenter.Name {
		a.vmCloudProperties.Datacenter.Name = cloudProps.Datacenter.Name
	}

	if len(cloudProps.SshKeys) > 0 {
		a.vmCloudProperties.SshKeys = cloudProps.SshKeys
	}
}
