package vm

import (
	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

type VMCloudProperties struct {
	VmNamePrefix             string                               `json:"vmNamePrefix,omitempty"`
	Domain                   string                               `json:"domain,omitempty"`
	StartCpus                int                                  `json:"startCpus,omitempty"`
	MaxMemory                int                                  `json:"maxMemory,omitempty"`
	Datacenter               sldatatypes.Datacenter               `json:"datacenter"`
	BlockDeviceTemplateGroup sldatatypes.BlockDeviceTemplateGroup `json:"blockDeviceTemplateGroup,omitempty"`
	SshKeys                  []sldatatypes.SshKey                 `json:"sshKeys,omitempty"`
	RootDiskSize             int                                  `json:"rootDiskSize,omitempty"`
	EphemeralDiskSize        int                                  `json:"ephemeralDiskSize,omitempty"`

	HourlyBillingFlag              bool                                       `json:"hourlyBillingFlag,omitempty"`
	LocalDiskFlag                  bool                                       `json:"localDiskFlag,omitempty"`
	DedicatedAccountHostOnlyFlag   bool                                       `json:"dedicatedAccountHostOnlyFlag,omitempty"`
	NetworkComponents              []sldatatypes.NetworkComponents            `json:"networkComponents,omitempty"`
	PrivateNetworkOnlyFlag         bool                                       `json:"privateNetworkOnlyFlag,omitempty"`
	PrimaryNetworkComponent        sldatatypes.PrimaryNetworkComponent        `json:"primaryNetworkComponent,omitempty"`
	PrimaryBackendNetworkComponent sldatatypes.PrimaryBackendNetworkComponent `json:"primaryBackendNetworkComponent,omitempty"`
	BlockDevices                   []sldatatypes.BlockDevice                  `json:"blockDevices,omitempty"`
	UserData                       []sldatatypes.UserData                     `json:"userData,omitempty"`
	PostInstallScriptUri           string                                     `json:"postInstallScriptUri,omitempty"`

	BoshIp string `json:"bosh_ip,omitempty"`
}

type AllowedHostCredential struct {
	Iqn      string `json:"iqn"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type VMMetadata map[string]interface{}

type Creator interface {
	// Create takes an agent id and creates a VM with provided configuration
	Create(string, bslcstem.Stemcell, VMCloudProperties, Networks, Environment) (VM, error)
}

type Finder interface {
	Find(int) (VM, bool, error)
}

type VM interface {
	ID() int

	Delete(agentId string) error
	Reboot() error

	SetMetadata(VMMetadata) error
	ConfigureNetworks(Networks) error

	AttachDisk(bslcdisk.Disk) error
	DetachDisk(bslcdisk.Disk) error
}

type Environment map[string]interface{}

type Mount struct {
	PartitionPath string
	MountPoint    string
}
