package action

import (
	"strconv"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

type CreateVM struct {
	stemcellFinder    bslcstem.Finder
	vmCreator         bslcvm.Creator
	vmCloudProperties *bslcvm.VMCloudProperties
}

type Environment map[string]interface{}

func NewCreateVM(stemcellFinder bslcstem.Finder, vmCreator bslcvm.Creator) CreateVM {
	return CreateVM{
		stemcellFinder:    stemcellFinder,
		vmCreator:         vmCreator,
		vmCloudProperties: &bslcvm.VMCloudProperties{},
	}
}

func (a CreateVM) Run(agentID string, stemcellCID StemcellCID, cloudProps bslcvm.VMCloudProperties, networks Networks, diskIDs []DiskCID, env Environment) (string, error) {
	a.UpdateCloudProperties(&cloudProps)

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

func (a CreateVM) UpdateCloudProperties(cloudProps *bslcvm.VMCloudProperties) {

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
