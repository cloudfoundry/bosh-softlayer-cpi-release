package action

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "bosh-softlayer-cpi/softlayer/common"
	bslnet "bosh-softlayer-cpi/softlayer/networks"
	bslstem "bosh-softlayer-cpi/softlayer/stemcell"
)

type CreateVMAction struct {
	stemcellFinder    bslstem.StemcellFinder
	vmCreatorProvider CreatorProvider
	vmCreator         VMCreator
	vmCloudProperties *VMCloudProperties
	options           ConcreteFactoryOptions
}

func NewCreateVM(
	stemcellFinder bslstem.StemcellFinder,
	vmCreatorProvider CreatorProvider,
	options ConcreteFactoryOptions,
) (action CreateVMAction) {
	action.options = options
	action.stemcellFinder = stemcellFinder
	action.vmCreatorProvider = vmCreatorProvider
	action.vmCloudProperties = &VMCloudProperties{}
	return
}

func (cvm CreateVMAction) Run(agentID string, stemcellCID StemcellCID, cloudProps VMCloudProperties, networks bslnet.Networks, diskIDs []DiskCID, env Environment) (string, error) {
	cvm.updateCloudProperties(&cloudProps)

	stemcell, err := cvm.stemcellFinder.FindById(int(stemcellCID))
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Finding stemcell '%s'", stemcellCID)
	}

	cvm.vmCreator = cvm.vmCreatorProvider.Get("virtualguest")
	vm, err := cvm.vmCreator.Create(agentID, stemcell, cloudProps, networks, env)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating Virtual_Guest with agent ID '%s'", agentID)
	}

	return VMCID(*vm.ID()).String(), nil

}

func (cvm CreateVMAction) updateCloudProperties(cloudProps *VMCloudProperties) {
	cvm.vmCloudProperties = cloudProps

	if cloudProps.DeployedByBoshCLI {
		cvm.vmCloudProperties.VmNamePrefix = updateHostNameInCloudProps(cloudProps, "")
	} else {
		cvm.vmCloudProperties.VmNamePrefix = updateHostNameInCloudProps(cloudProps, TimeStampForTime(time.Now().UTC()))
	}

	if cloudProps.StartCpus == 0 {
		cvm.vmCloudProperties.StartCpus = 4
	}

	if cloudProps.MaxMemory == 0 {
		cvm.vmCloudProperties.MaxMemory = 8192
	}

	if len(cloudProps.Domain) == 0 {
		cvm.vmCloudProperties.Domain = "softlayer.com"
	}

	if cloudProps.MaxNetworkSpeed == 0 {
		cvm.vmCloudProperties.MaxNetworkSpeed = 1000
	}
}

func updateHostNameInCloudProps(cloudProps *VMCloudProperties, timeStampPostfix string) string {
	if len(timeStampPostfix) == 0 {
		return cloudProps.VmNamePrefix
	} else {
		return cloudProps.VmNamePrefix + "." + timeStampPostfix
	}
}
