package action

import (
	"strconv"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

type createVM struct {
	stemcellFinder    bslcstem.Finder
	vmCreatorProvider Provider
	vmCloudProperties *bslcvm.VMCloudProperties
}

type Environment map[string]interface{}

func NewCreateVM(stemcellFinder bslcstem.Finder, vmCreatorProvider Provider) Action {
	return &createVM{
		stemcellFinder:    stemcellFinder,
		vmCreatorProvider: vmCreatorProvider,
		vmCloudProperties: &bslcvm.VMCloudProperties{},
	}
}

func (a *createVM) Run(agentID string, stemcellCID StemcellCID, cloudProps bslcvm.VMCloudProperties, networks Networks, diskIDs []DiskCID, env Environment) (string, error) {
	vmNetworks := networks.AsVMNetworks()
	vmEnv := bslcvm.Environment(env)

	a.UpdateCloudProperties(&cloudProps)

	stemcell, found, err := a.stemcellFinder.FindById(int(stemcellCID))
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Finding stemcell '%s'", stemcellCID)
	}

	if !found {
		return "", bosherr.Errorf("Expected to find stemcell '%s'", stemcellCID)
	}

	if cloudProps.Baremetal {
		vmCreator, err := a.vmCreatorProvider.Get("baremetal")
		if err != nil {
			return "", bosherr.WrapError(err, "Failed to get baremetal creator'")
		}

		vm, err := vmCreator.Create(agentID, stemcell, cloudProps, vmNetworks, vmEnv)
		if err != nil {
			return "", bosherr.WrapErrorf(err, "Creating Baremetal with agent ID '%s'", agentID)
		}
		return VMCID(vm.ID()).String(), nil
	} else {
		vmCreator, err := a.vmCreatorProvider.Get("virtualguest")
		if err != nil {
			return "", bosherr.WrapError(err, "Failed to get virtual_guest creator'")
		}

		vm, err := vmCreator.Create(agentID, stemcell, cloudProps, vmNetworks, vmEnv)
		if err != nil {
			return "", bosherr.WrapErrorf(err, "Creating Virtual_Guest with agent ID '%s'", agentID)
		}
		return VMCID(vm.ID()).String(), nil
	}
}

func (a *createVM) UpdateCloudProperties(cloudProps *bslcvm.VMCloudProperties) {

	a.vmCloudProperties = cloudProps

	if len(cloudProps.BoshIp) == 0 {
		a.vmCloudProperties.VmNamePrefix = cloudProps.VmNamePrefix
	} else {
		a.vmCloudProperties.VmNamePrefix = cloudProps.VmNamePrefix + timeStampForTime(time.Now().UTC())
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

func timeStampForTime(now time.Time) string {
	//utilize the constants list in the http://golang.org/src/time/format.go file to get the expect time formats
	return now.Format("20060102-030405-") + strconv.Itoa(int(now.UnixNano()/1e6-now.Unix()*1e3))
}
