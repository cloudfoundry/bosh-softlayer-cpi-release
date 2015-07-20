package vm

import (
	"encoding/base64"
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

func (c SoftLayerCreator) getTimeStamp(now time.Time) string {
	//utilize the constants list in the http://golang.org/src/time/format.go file to get the expect time formats
	return now.Format("20060102-030405-") + strconv.Itoa(int(now.UnixNano()/1e6-now.Unix()*1e3))
}

func (c SoftLayerCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	disks := DisksSpec{}
	if cloudProps.EphemeralDiskSize > 0 {
		disks = DisksSpec{Ephemeral: "/dev/xvdc"}
	}
	// To keep consistent with legacy softlayer CPI, making agent name with prefix "vm-"
	agentEnv := NewAgentEnvForVM(agentID, fmt.Sprintf("vm-%s", agentID), networks, disks, env, c.agentOptions)
	metadata, err := json.Marshal(agentEnv)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Marshalling agent environment metadata")
	}
	dataBytes := []byte(metadata)
	base64EncodedMetadata := base64.StdEncoding.EncodeToString(dataBytes)

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
		LocalDiskFlag:     cloudProps.LocalDiskFlag,

		DedicatedAccountHostOnlyFlag:   cloudProps.DedicatedAccountHostOnlyFlag,
		BlockDevices:                   cloudProps.BlockDevices,
		NetworkComponents:              cloudProps.NetworkComponents,
		PrivateNetworkOnlyFlag:         cloudProps.PrivateNetworkOnlyFlag,
		PrimaryNetworkComponent:        &cloudProps.PrimaryNetworkComponent,
		PrimaryBackendNetworkComponent: &cloudProps.PrimaryBackendNetworkComponent,
		UserData: []sldatatypes.UserData{
			sldatatypes.UserData{
				Value: base64EncodedMetadata,
			},
		},
	}

	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	virtualGuest, err := virtualGuestService.CreateObject(virtualGuestTemplate)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating VirtualGuest from SoftLayer client")
	}

	if cloudProps.EphemeralDiskSize == 0 {
		err = bslcommon.WaitForVirtualGuest(c.softLayerClient, virtualGuest.Id, "RUNNING", bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("PowerOn failed with VirtualGuest id `%d`", virtualGuest.Id))
		}
	} else {
		err = bslcommon.AttachEphemeralDiskToVirtualGuest(c.softLayerClient, virtualGuest.Id, cloudProps.EphemeralDiskSize, bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("Attaching ephemeral disk to VirtualGuest `%d`", virtualGuest.Id))
		}
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
