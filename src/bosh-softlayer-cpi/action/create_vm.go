package action

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "bosh-softlayer-cpi/softlayer/common"
	bslcstem "bosh-softlayer-cpi/softlayer/stemcell"

	"bosh-softlayer-cpi/api"
	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

type CreateVMAction struct {
	stemcellFinder    bslcstem.StemcellFinder
	vmCreatorProvider CreatorProvider
	vmCreator         VMCreator
	vmCloudProperties *VMCloudProperties
	options           ConcreteFactoryOptions
}

func NewCreateVM(
	stemcellFinder bslcstem.StemcellFinder,
	vmCreatorProvider CreatorProvider,
	options ConcreteFactoryOptions,
) (action CreateVMAction) {
	action.options = options
	action.stemcellFinder = stemcellFinder
	action.vmCreatorProvider = vmCreatorProvider
	action.vmCloudProperties = &VMCloudProperties{}
	return
}

func (a CreateVMAction) Run(agentID string, stemcellCID StemcellCID, cloudProps VMCloudProperties, networks Networks, diskIDs []DiskCID, env Environment) (string, error) {
	a.updateCloudProperties(&cloudProps)

	api.TIMEOUT = 30 * time.Second
	api.POLLING_INTERVAL = 5 * time.Second
	api.NetworkInterface = "eth0"
	api.LocalDNSConfigurationFile = "/etc/hosts"

	stemcell, err := a.stemcellFinder.FindById(int(stemcellCID))
	if err != nil {
		return "0", bosherr.WrapErrorf(err, "Finding stemcell '%s'", stemcellCID)
	}

	if a.options.Softlayer.FeatureOptions.EnablePool {
		a.vmCreator = a.vmCreatorProvider.Get("pool")
		vm, err := a.vmCreator.Create(agentID, stemcell, cloudProps, networks, env)
		if err != nil {
			return "0", bosherr.WrapErrorf(err, "Creating vm with agent ID '%s'", agentID)
		}

		return VMCID(vm.ID()).String(), nil
	}

	if cloudProps.Baremetal {
		a.vmCreator = a.vmCreatorProvider.Get("baremetal")
		vm, err := a.vmCreator.Create(agentID, stemcell, cloudProps, networks, env)
		if err != nil {
			return "0", bosherr.WrapErrorf(err, "Creating Baremetal with agent ID '%s'", agentID)
		}

		return VMCID(vm.ID()).String(), nil
	} else {
		a.vmCreator = a.vmCreatorProvider.Get("virtualguest")
		vm, err := a.vmCreator.Create(agentID, stemcell, cloudProps, networks, env)
		if err != nil {
			return "0", bosherr.WrapErrorf(err, "Creating Virtual_Guest with agent ID '%s'", agentID)
		}

		return VMCID(vm.ID()).String(), nil
	}
}

func (a CreateVMAction) updateCloudProperties(cloudProps *VMCloudProperties) {
	a.vmCloudProperties = cloudProps

	if cloudProps.DeployedByBoshCLI {
		a.vmCloudProperties.VmNamePrefix = cloudProps.VmNamePrefix
	} else {
		a.vmCloudProperties.VmNamePrefix = cloudProps.VmNamePrefix + TimeStampForTime(time.Now().UTC())
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
	api.LengthOfHostName = len(a.vmCloudProperties.VmNamePrefix + "." + a.vmCloudProperties.Domain)
	// A workaround for the issue #129 in bosh-softlayer-cpi
	if api.LengthOfHostName == 64 {
		a.vmCloudProperties.VmNamePrefix = a.vmCloudProperties.VmNamePrefix + "-1"
		api.LengthOfHostName = len(a.vmCloudProperties.VmNamePrefix + "." + a.vmCloudProperties.Domain)
	}
	if len(cloudProps.NetworkComponents) == 0 {
		a.vmCloudProperties.NetworkComponents = []sldatatypes.NetworkComponents{{MaxSpeed: 1000}}
	}

	if api.LocalDiskFlagNotSet == true {
		a.vmCloudProperties.LocalDiskFlag = true
	}
}
