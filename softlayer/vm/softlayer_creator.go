package vm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcommon "github.com/maximilien/bosh-softlayer-cpi/softlayer/common"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	util "github.com/maximilien/bosh-softlayer-cpi/util"
)

const softLayerCreatorLogTag = "SoftLayerCreator"

type SoftLayerCreator struct {
	softLayerClient        sl.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	agentOptions AgentOptions
	logger       boshlog.Logger
}

func NewSoftLayerCreator(softLayerClient sl.Client, agentEnvServiceFactory AgentEnvServiceFactory, agentOptions AgentOptions, logger boshlog.Logger) SoftLayerCreator {
	bslcommon.TIMEOUT = 20 * time.Minute
	bslcommon.POLLING_INTERVAL = 20 * time.Second
	bslcommon.MAX_RETRY_COUNT = 5

	return SoftLayerCreator{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		agentOptions:           agentOptions,
		logger:                 logger,
	}
}
// get the time stample of action occurs to name the vm hostname
func (c SoftLayerCreator) getTimeStample(now time.Time) string {
	return now.Format("20060102-030405-") + strconv.Itoa(int(now.UnixNano()/1e6- now.Unix() * 1e3))
}

func (c SoftLayerCreator) checkValid(t *sldatatypes.SoftLayer_Virtual_Guest_Template)(*sldatatypes.SoftLayer_Virtual_Guest_Template){
	//check domain name configure or not
	if t.Domain =="" {
		t.Domain = "softlayer.com"
	}
	//check conditional require
	if t.BlockDeviceTemplateGroup.GlobalIdentifier != "" {
		t.OperatingSystemReferenceCode =""
		t.BlockDevices = nil
	}

	return t
}

func (c SoftLayerCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {

	virtualGuestTemplate := sldatatypes.SoftLayer_Virtual_Guest_Template{
		//Hostname:  agentID,
		Hostname:  cloudProps.VMNamePrefix + c.getTimeStample(time.Now().UTC()),
		Domain:    cloudProps.Domain,
		StartCpus: cloudProps.StartCpus,
		MaxMemory: cloudProps.MaxMemory,

		Datacenter: sldatatypes.Datacenter{
			Name: cloudProps.Datacenter.Name,
		},

		BlockDeviceTemplateGroup: &sldatatypes.BlockDeviceTemplateGroup{
			GlobalIdentifier: stemcell.Uuid(),
		},

		SshKeys:           cloudProps.SshKeys,

		//HourlyBillingFlag: true,
		HourlyBillingFlag: cloudProps.HourlyBillingFlag,
		//Needed for ephemeral disk
		//LocalDiskFlag: true,
		LocalDiskFlag: cloudProps.LocalDiskFlag,
		DedicatedAccountHostOnlyFlag: cloudProps.DedicatedAccountHostOnlyFlag,
		BlockDevices:cloudProps.BlockDevices,
		NetworkComponents : cloudProps.NetworkComponents,
		PrivateNetworkOnlyFlag: cloudProps.PrivateNetworkOnlyFlag,
		PrimaryNetworkComponent:cloudProps.PrimaryNetworkComponent,
		PrimaryBackendNetworkComponent: cloudProps.PrimaryBackendNetworkComponent,
		UserData:cloudProps.UserData,
	}

	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	virtualGuest, err := virtualGuestService.CreateObject(*c.checkValid(&virtualGuestTemplate))
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating VirtualGuest from SoftLayer client")
	}

	//TODO: need to find or ensure the name for the ephemeral disk for SoftLayer VG
	disks := DisksSpec{Ephemeral: "/dev/xvdc"}

	agentEnv := NewAgentEnvForVM(agentID, strconv.Itoa(virtualGuest.Id), networks, disks, env, c.agentOptions)

	metadata, err := json.Marshal(agentEnv)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Marshalling agent environment metadata")
	}

	err = bslcommon.ConfigureMetadataOnVirtualGuest(c.softLayerClient, virtualGuest.Id, string(metadata), bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("Configuring metadata on VirtualGuest `%d`", virtualGuest.Id))
	}

	err = bslcommon.AttachEphemeralDiskToVirtualGuest(c.softLayerClient, virtualGuest.Id, cloudProps.EphemeralDiskSize, bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("Attaching ephemeral disk to VirtualGuest `%d`", virtualGuest.Id))
	}

	agentEnvService := c.agentEnvServiceFactory.New(virtualGuest.Id)

	vm := NewSoftLayerVM(virtualGuest.Id, c.softLayerClient, util.GetSshClient(), agentEnvService, c.logger)

	return vm, nil
}

func (c SoftLayerCreator) resolveNetworkIP(networks Networks) (string, error) {
	var network Network

	switch len(networks) {
	case 0:
		return "", bosherr.Error("Expected exactly one network; received zero")
	case 1:
		network = networks.First()
	default:
		return "", bosherr.Error("Expected exactly one network; received multiple")
	}

	if network.IsDynamic() {
		return "", nil
	}

	return network.IP, nil
}
