package action

import (
	"encoding/json"

	instance "bosh-softlayer-cpi/softlayer/virtual_guest_service"
	"fmt"
	"time"
)

type DiskCloudProperties struct {
	DiskType         string `json:"type,omitempty"`
	DataCenter       string `json:"datacenter,omitempty"`
	Iops             int    `json:"iops,omitempty"`
	SnapShotSpace    int    `json:"snapshotSpace,omitempty"`
	UseHourlyPricing bool   `json:"useHourlyPricing,omitempty"`
}

type Environment map[string]interface{}

type NetworkCloudProperties struct {
	VlanIds             []int         `json:"vlanIds,omitempty"`
	SourcePolicyRouting bool          `json:"source_policy_routing,omitempty"`
	Tags                instance.Tags `json:"tags,omitempty"`
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

	// URL of an existing image (Image.SelfLink)
	ImageURL   string `json:"image_url,omitempty"`
	SourceSha1 string `json:"raw_disk_sha1,omitempty"`
}

type VMCloudProperties struct {
	VmNamePrefix      string `json:"vmNamePrefix,omitempty"`
	Domain            string `json:"domain,omitempty"`
	StartCpus         int    `json:"startCpus,omitempty"`
	MaxMemory         int    `json:"maxMemory,omitempty"`
	Datacenter        string `json:"dataCenter"`
	EphemeralDiskSize int    `json:"ephemeralDiskSize,omitempty"`
	SshKey            int    `json:"sshKey,omitempty"`

	HourlyBillingFlag            bool `json:"hourlyBillingFlag,omitempty"`
	LocalDiskFlag                bool `json:"localDiskFlag,omitempty"`
	DedicatedAccountHostOnlyFlag bool `json:"dedicatedAccountHostOnlyFlag,omitempty"`
	PrivateNetworkOnlyFlag       bool `json:"privateNetworkOnlyFlag,omitempty"`

	DeployedByBoshCLI bool `json:"deployedByBoshCli,omitempty"`

	MaxNetworkSpeed int `json:"maxNetworkSpeed,omitempty"`

	Tags instance.Tags `json:"tags,omitempty"`
}

func (vmProps *VMCloudProperties) Validate() error {
	if err := vmProps.Tags.Validate(); err != nil {
		return err
	}

	return nil
}

func (vmProps *VMCloudProperties) AsInstanceProperties() *VMCloudProperties {
	if vmProps.DeployedByBoshCLI {
		vmProps.VmNamePrefix = vmProps.updateHostNameInCloudProps(vmProps, "")
	} else {
		vmProps.VmNamePrefix = vmProps.updateHostNameInCloudProps(vmProps, timeStampForTime(time.Now().UTC()))
	}

	if vmProps.StartCpus == 0 {
		vmProps.StartCpus = 4
	}

	if vmProps.MaxMemory == 0 {
		vmProps.MaxMemory = 8192
	}

	if len(vmProps.Domain) == 0 {
		vmProps.Domain = "softlayer.com"
	}

	if vmProps.MaxNetworkSpeed == 0 {
		vmProps.MaxNetworkSpeed = 1000
	}

	return vmProps
}

func (vmProps VMCloudProperties) updateHostNameInCloudProps(cloudProps *VMCloudProperties, timeStampPostfix string) string {
	if len(timeStampPostfix) == 0 {
		return cloudProps.VmNamePrefix
	} else {
		return cloudProps.VmNamePrefix + "-" + timeStampPostfix
	}
}

func timeStampForTime(now time.Time) string {
	//utilize the constants list in the http://golang.org/src/time/format.go file to get the expect time formats
	return now.Format("20060102-030405-") + fmt.Sprintf("%03d", int(now.UnixNano()/1e6-now.Unix()*1e3))
}

type VMMetadata map[string]interface{}
