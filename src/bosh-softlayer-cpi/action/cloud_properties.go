package action

import (
	"encoding/json"
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DiskCloudProperties struct {
	DiskType      string `json:"type,omitempty"`
	DataCenter    string `json:"datacenter,omitempty"`
	Iops          int    `json:"iops,omitempty"`
	SnapshotSpace int    `json:"snapshot_space,omitempty"`
}

type Environment map[string]interface{}

type NetworkCloudProperties struct {
	SubnetIds           []int `json:"subnet_ids,omitempty"`
	VlanIds             []int `json:"vlan_ids,omitempty"`
	SourcePolicyRouting bool  `json:"source_policy_routing,omitempty"`
}

type SnapshotMetadata struct {
	Deployment string      `json:"deployment,omitempty"`
	Job        string      `json:"job,omitempty"`
	Index      json.Number `json:"index,omitempty"`
}

type StemcellCloudProperties struct {
	Name           string `json:"name,omitempty"`
	Version        string `json:"version,omitempty"`
	Infrastructure string `json:"infrastructure,omitempty"`
	SourceURL      string `json:"source_url,omitempty"`

	//// URL of an existing image (Image.SelfLink)
	//ImageURL   string `json:"image_url,omitempty"`
	//SourceSha1 string `json:"raw_disk_sha1,omitempty"`

	Id             int    `json:"virtual-disk-image-id"`
	Uuid           string `json:"virtual-disk-image-uuid"`

	DatacenterName string `json:"datacenter-name"`
	OsCode         string `json:"os-code"`
}

type VMCloudProperties struct {
	HostnamePrefix    string `json:"hostname_prefix,omitempty"`
	Hostname          string `json:"hostname,omitempty"`
	Domain            string `json:"domain,omitempty"`
	FlavorKeyName     string `json:"flavor_key_name,omitempty"`
	Cpu               int    `json:"cpu,omitempty"`
	Memory            int    `json:"memory,omitempty"`
	Datacenter        string `json:"datacenter"`
	EphemeralDiskSize int    `json:"ephemeral_disk_size,omitempty"`
	SshKey            int    `json:"ssh_key,omitempty"`

	HourlyBillingFlag            bool `json:"hourly_billing_flag,omitempty"`
	LocalDiskFlag                bool `json:"local_disk_flag,omitempty"`
	DedicatedAccountHostOnlyFlag bool `json:"dedicated_account_host_only_flag,omitempty"`

	DeployedByBoshCLI bool `json:"deployed_by_boshcli,omitempty"`

	MaxNetworkSpeed int `json:"max_network_speed,omitempty"`
}

func (vmProps *VMCloudProperties) Validate() error {
	if vmProps.HostnamePrefix == "" {
		return bosherr.Error("The property 'hostname_prefix' must be set to create an instance")
	}
	if vmProps.Domain == "" {
		vmProps.Domain = "softlayer.com"
	}
	if vmProps.Datacenter == "" {
		return bosherr.Error("The property 'datacenter' must be set to create an instance")
	}

	if vmProps.FlavorKeyName != "" && (vmProps.Memory != 0 || vmProps.Cpu != 0) {
		return bosherr.Error("The property 'flavor_key_name' can not be set with 'memory/cpu'")
	} else {
		if vmProps.Memory == 0 {
			vmProps.Memory = 8192
		}
		if vmProps.Cpu == 0 {
			vmProps.Cpu = 4
		}
	}

	if vmProps.MaxNetworkSpeed == 0 {
		vmProps.MaxNetworkSpeed = 1000
	}

	return nil
}

func (vmProps *VMCloudProperties) AsInstanceProperties() *VMCloudProperties {
	if vmProps.DeployedByBoshCLI {
		vmProps.Hostname = vmProps.updateHostNameInCloudProps(vmProps, "")
	} else {
		vmProps.Hostname = vmProps.updateHostNameInCloudProps(vmProps, timeStampForTime(time.Now().UTC()))
	}

	// A workaround for the issue #129 in bosh-softlayer-cpi
	if len(vmProps.Hostname+"."+vmProps.Domain) == 64 {
		vmProps.Hostname = vmProps.Hostname + "-1"
	}

	return vmProps
}

func (vmProps VMCloudProperties) updateHostNameInCloudProps(cloudProps *VMCloudProperties, timeStampPostfix string) string {
	if len(timeStampPostfix) == 0 {
		return cloudProps.HostnamePrefix
	} else {
		prefixLen := len(cloudProps.HostnamePrefix)
		if prefixLen > 0 && cloudProps.HostnamePrefix[prefixLen-1] == '-' {
			return cloudProps.HostnamePrefix + timeStampPostfix
		}
		return cloudProps.HostnamePrefix + "-" + timeStampPostfix
	}
}

func timeStampForTime(now time.Time) string {
	//utilize the constants list in the http://golang.org/src/time/format.go file to get the expect time formats
	return now.Format("20060102-030405-") + fmt.Sprintf("%03d", int(now.UnixNano()/1e6-now.Unix()*1e3))
}

type VMMetadata map[string]interface{}

type DiskMetadata map[string]interface{}
