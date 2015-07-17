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
	"strings"
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

func (c SoftLayerCreator) getTimeStamp(now time.Time) string {
	//utilize the constants list in the http://golang.org/src/time/format.go file to get the expect time formats
	return now.Format("20060102-030405-") + strconv.Itoa(int(now.UnixNano()/1e6-now.Unix()*1e3))
}

func (c SoftLayerCreator) CreateNewVM(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {

	c.logger.Info(softLayerCreatorLogTag, fmt.Sprintln("Creating a new server with stemcell %s",  stemcell.ID()))

	virtualGuestTemplate := sldatatypes.SoftLayer_Virtual_Guest_Template{
		Hostname:  cloudProps.VmNamePrefix + c.getTimeStamp(time.Now().UTC()),
		Domain:    cloudProps.Domain,
		StartCpus: cloudProps.StartCpus,
		MaxMemory: cloudProps.MaxMemory,

		Datacenter: sldatatypes.Datacenter{
			Name: cloudProps.Datacenter.Name,
		},

		BlockDeviceTemplateGroup: &sldatatypes.BlockDeviceTemplateGroup{
			GlobalIdentifier: stemcell.Uuid(),
		},

		SshKeys: cloudProps.SshKeys,

		HourlyBillingFlag: cloudProps.HourlyBillingFlag,
		//Needed for ephemeral disk
		LocalDiskFlag: cloudProps.LocalDiskFlag,

		DedicatedAccountHostOnlyFlag:   cloudProps.DedicatedAccountHostOnlyFlag,
		BlockDevices:                   cloudProps.BlockDevices,
		NetworkComponents:              cloudProps.NetworkComponents,
		PrivateNetworkOnlyFlag:         cloudProps.PrivateNetworkOnlyFlag,
		PrimaryNetworkComponent:        &cloudProps.PrimaryNetworkComponent,
		PrimaryBackendNetworkComponent: &cloudProps.PrimaryBackendNetworkComponent,
		UserData:                       cloudProps.UserData,
	}

	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	virtualGuest, err := virtualGuestService.CreateObject(virtualGuestTemplate)
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

	//Insert a record into VM pool table (sqlite DB) when creating a new VM
	vmInfo := VMInfo{
		id:vm.id,
		name:virtualGuestTemplate.Hostname+virtualGuestTemplate.Domain,
		in_use:"t",
		image_id:stemcell.Uuid(),
		agent_id:agentID,
	}
	err = InsertVMInfo(&vmInfo)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Failed to insert the record into VM pool DB")
	}

	return vm, nil

}

func (c SoftLayerCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {

	if strings.Contains(cloudProps.VmNamePrefix, "-worker") {
		vm, err := c.CreateNewVM(agentID, stemcell, cloudProps, networks, env)
		return vm, err
	}

	err := InitVMPoolDB()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Failed to initialize VM pool DB")
	}

	vmInfo, err := QueryVMInfobyAgentID(agentID)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "")
	}

	if vmInfo.id != 0 {
		c.logger.Info(softLayerCreatorLogTag, fmt.Sprintln("OS reload on the server id %d with stemcell %s", vmInfo.id,  stemcell))
		agentEnvService := c.agentEnvServiceFactory.New(vmInfo.id)
		vm := NewSoftLayerVM(vmInfo.id, c.softLayerClient, util.GetSshClient(), agentEnvService, c.logger)
		vm.ReloadOS(stemcell)
		return vm, nil
	}

	return c.CreateNewVM(agentID, stemcell, cloudProps, networks, env)

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
