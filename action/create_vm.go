package action

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	"github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"

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

	helper.TIMEOUT = 30 * time.Second
	helper.POLLING_INTERVAL = 5 * time.Second
	helper.NetworkInterface = "eth0"
	helper.LocalDNSConfigurationFile = "/etc/hosts"

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

	if len(cloudProps.BoshIp) == 0 || cloudProps.Baremetal {
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
	helper.LengthOfHostName = len(a.vmCloudProperties.VmNamePrefix + "." + a.vmCloudProperties.Domain)

	if len(cloudProps.NetworkComponents) == 0 {
		a.vmCloudProperties.NetworkComponents = []sldatatypes.NetworkComponents{{MaxSpeed: 1000}}
	}

	if helper.LocalDiskFlagNotSet == true {
		a.vmCloudProperties.LocalDiskFlag = true
	}
}
