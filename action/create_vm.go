package action

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

type CreateVMAction struct {
	stemcellFinder    bslcstem.Finder
	vmCreatorProvider Provider
	vmCreator         bslcvm.VMCreator
	vmCloudProperties *bslcvm.VMCloudProperties
}

type Environment map[string]interface{}

func NewCreateVM(
	stemcellFinder bslcstem.Finder,
	vmCreatorProvider Provider,
) (action CreateVMAction) {
	action.stemcellFinder = stemcellFinder
	action.vmCreatorProvider = vmCreatorProvider
	action.vmCloudProperties = &bslcvm.VMCloudProperties{}
	return
}

func (a CreateVMAction) Run(agentID string, stemcellCID StemcellCID, cloudProps bslcvm.VMCloudProperties, networks Networks, diskIDs []DiskCID, env Environment) (string, error) {
	vmNetworks := networks.AsVMNetworks()
	vmEnv := bslcvm.Environment(env)

	a.UpdateCloudProperties(&cloudProps)

	stemcell, found, err := a.stemcellFinder.FindById(int(stemcellCID))
	if err != nil {
		return "0", bosherr.WrapErrorf(err, "Finding stemcell '%s'", stemcellCID)
	}

	if !found {
		return "0", bosherr.Errorf("Expected to find stemcell '%s'", stemcellCID)
	}

	if cloudProps.Baremetal {
		a.vmCreator, err = a.vmCreatorProvider.Get("baremetal")
		if err != nil {
			return "0", bosherr.WrapError(err, "Failed to get baremetal creator'")
		}

		vm, err := a.vmCreator.Create(agentID, stemcell, cloudProps, vmNetworks, vmEnv)
		if err != nil {
			return "0", bosherr.WrapErrorf(err, "Creating Baremetal with agent ID '%s'", agentID)
		}
		return VMCID(vm.ID()).String(), nil
	} else {
		a.vmCreator, err = a.vmCreatorProvider.Get("virtualguest")
		if err != nil {
			return "0", bosherr.WrapError(err, "Failed to get virtual_guest creator'")
		}

		vm, err := a.vmCreator.Create(agentID, stemcell, cloudProps, vmNetworks, vmEnv)
		if err != nil {
			return "0", bosherr.WrapErrorf(err, "Creating Virtual_Guest with agent ID '%s'", agentID)
		}
		return VMCID(vm.ID()).String(), nil
	}
}

func (a CreateVMAction) UpdateCloudProperties(cloudProps *bslcvm.VMCloudProperties) {
	a.vmCloudProperties = cloudProps

	if len(cloudProps.BoshIp) == 0 {
		a.vmCloudProperties.VmNamePrefix = cloudProps.VmNamePrefix
	} else {
		a.vmCloudProperties.VmNamePrefix = cloudProps.VmNamePrefix + bslcvm.TimeStampForTime(time.Now().UTC())
	}

	if cloudProps.StartCpus == 0 {
		a.vmCloudProperties.StartCpus = 4
	}

	if cloudProps.MaxMemory == 0 {
		a.vmCloudProperties.MaxMemory = 8192
	}

	if len(cloudProps.Domain) == 0 {
		a.vmCloudProperties.Domain = "softlayer.com"
	}
	if len(cloudProps.NetworkComponents) == 0 {
		a.vmCloudProperties.NetworkComponents = []sldatatypes.NetworkComponents{{MaxSpeed: 1000}}
	}
}
