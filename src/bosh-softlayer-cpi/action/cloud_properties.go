package action

import (
	"encoding/json"

	instance "bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

type DiskCloudProperties struct {
	DiskType         string `json:"type,omitempty"`
	DataCenter       string `json:"datacenter,omitempty"`
	Iops             int    `json:"iops,omitempty"`
	UseHourlyPricing bool   `json:"useHourlyPricing,omitempty"`
}

type Environment map[string]interface{}

type NetworkCloudProperties struct {
	VlanID              int           `json:"vlanId"`
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
	Datacenter        string `json:"datacenter"`
	EphemeralDiskSize int    `json:"ephemeralDiskSize,omitempty"`
	SshKey            int    `json:"sshKey,omitempty"`

	HourlyBillingFlag            bool `json:"hourlyBillingFlag,omitempty"`
	LocalDiskFlag                bool `json:"localDiskFlag,omitempty"`
	DedicatedAccountHostOnlyFlag bool `json:"dedicatedAccountHostOnlyFlag,omitempty"`
	PrivateNetworkOnlyFlag       bool `json:"privateNetworkOnlyFlag,omitempty"`

	DeployedByBoshCLI bool `json:"deployedByBoshcli,omitempty"`

	MaxNetworkSpeed int `json:"maxNetworkSpeed,omitempty"`

	Tags instance.Tags `json:"tags,omitempty"`
}

func (n VMCloudProperties) Validate() error {
	if err := n.Tags.Validate(); err != nil {
		return err
	}

	return nil
}

type VMMetadata map[string]interface{}
