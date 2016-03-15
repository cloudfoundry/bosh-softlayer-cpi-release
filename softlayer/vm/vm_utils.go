package vm

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"

	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
)

func CreateDisksSpec(ephemeralDiskSize int) DisksSpec {
	disks := DisksSpec{}
	if ephemeralDiskSize > 0 {
		disks = DisksSpec{
			Ephemeral:  "/dev/xvdc",
			Persistent: nil,
		}
	}

	return disks
}

func TimeStampForTime(now time.Time) string {
	//utilize the constants list in the http://golang.org/src/time/format.go file to get the expect time formats
	return now.Format("20060102-030405-") + strconv.Itoa(int(now.UnixNano()/1e6-now.Unix()*1e3))
}

func CreateVirtualGuestTemplate(stemcell bslcstem.Stemcell, cloudProps VMCloudProperties) (sldatatypes.SoftLayer_Virtual_Guest_Template, error) {
	virtualGuestTemplate := sldatatypes.SoftLayer_Virtual_Guest_Template{
		Hostname:  cloudProps.VmNamePrefix,
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
	}

	return virtualGuestTemplate, nil
}

func CreateAgentUserData(agentID string, cloudProps VMCloudProperties, networks Networks, env Environment, agentOptions AgentOptions) AgentEnv {
	agentName := fmt.Sprintf("vm-%s", agentID)
	disks := CreateDisksSpec(cloudProps.EphemeralDiskSize)
	agentEnv := NewAgentEnvForVM(agentID, agentName, networks, disks, env, agentOptions)
	return agentEnv
}

func AppendPowerDNSToNetworks(networks Networks, cloudProps VMCloudProperties) Networks {
	powerdnsNetworks := make(map[string]Network)
	for name, network := range networks {
		network = network.AppendDNS(cloudProps.BoshIp)
		powerdnsNetworks[name] = network
	}

	return powerdnsNetworks
}

func Base64EncodeData(unEncodedData string) string {
	dataBytes := []byte(unEncodedData)
	return base64.StdEncoding.EncodeToString(dataBytes)
}
