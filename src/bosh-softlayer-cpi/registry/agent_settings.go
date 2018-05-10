package registry

import "encoding/json"

const DefaultEphemeralDisk = "/dev/xvdc"

type agentSettingsResponse struct {
	Settings string `json:"settings"`
	Status   string `json:"status"`
}

// AgentSettings are the Agent settings for a particular VM.
type AgentSettings struct {
	// Agent ID
	AgentID string `json:"agent_id"`

	// Blobstore settings
	Blobstore BlobstoreSettings `json:"blobstore"`

	// Disks settings
	Disks DisksSettings `json:"disks"`

	// Environment settings
	Env EnvSettings `json:"env"`

	// Mbus URI
	Mbus string `json:"mbus"`

	// Networks settings
	Networks NetworksSettings `json:"networks"`

	// List of NTP servers
	Ntp []string `json:"ntp"`

	// VM settings
	VM VMSettings `json:"vm"`
}

// BlobstoreSettings are the Blobstore settings for a particular VM.
type BlobstoreSettings struct {
	// Blobstore provider
	Provider string `json:"provider"`

	// Blobstore options
	Options map[string]interface{} `json:"options"`
}

// DisksSettings are the Disks settings for a particular VM.
type DisksSettings struct {
	// System disk
	System string `json:"system"`

	// Ephemeral disk
	Ephemeral string `json:"ephemeral"`

	// Persistent disk
	Persistent map[string]PersistentSettings `json:"persistent"`
}

// PersistentSettings are the Persistent Disk settings for a particular VM.
type PersistentSettings struct {
	// Persistent disk ID
	ID string `json:"id"`

	// For iscsi setup info
	ISCSISettings ISCSISettings `json:"iscsi_settings"`
}

type ISCSISettings struct {
	InitiatorName string `json:"initiator_name"`
	Target        string `json:"target"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

// EnvSettings are the Environment settings for a particular VM.
type EnvSettings map[string]interface{}

// NetworksSettings are the Networks settings for a particular VM.
type NetworksSettings map[string]NetworkSettings

// NetworkSettings are the Network settings for a particular VM.
type NetworkSettings struct {
	// Network type
	Type string `json:"type"`

	// Network IP address
	IP string `json:"ip"`

	// Network gateway
	Gateway string `json:"gateway"`

	// Network MAC address
	Mac string `json:"mac"`

	// Network netmask
	Netmask string `json:"netmask"`

	// List of DNS servers
	DNS []string `json:"dns"`

	// Does network have DHCP
	DHCP bool `json:"use_dhcp"`

	// List of defaults
	Default []string `json:"default"`

	Alias string `json:"alias,omitempty"`

	Routes Routes `json:"routes,omitempty"`

	// Does network is preconfigured
	Preconfigured bool `json:"preconfigured"`

	// Network cloud properties
	CloudProperties map[string]interface{} `json:"cloud_properties"`
}

type Routes []Route

type Route struct {
	Destination string
	Gateway     string
	NetMask     string
}

// VMSettings are the VM settings for a particular VM.
type VMSettings struct {
	// VM name
	Name string `json:"name"`
}

// NewAgentSettings creates new agent settings for a VM.
func NewAgentSettings(agentID string, vmCID string, networksSettings NetworksSettings, env EnvSettings, agentOptions AgentOptions) AgentSettings {
	agentSettings := AgentSettings{
		AgentID: agentID,
		Disks: DisksSettings{
			Ephemeral:  "",
			Persistent: map[string]PersistentSettings{},
		},
		Blobstore: BlobstoreSettings{
			Provider: agentOptions.Blobstore.Provider,
			Options:  agentOptions.Blobstore.Options,
		},
		Env:      EnvSettings(env),
		Mbus:     agentOptions.Mbus,
		Networks: networksSettings,
		Ntp:      agentOptions.Ntp,
		VM: VMSettings{
			Name: vmCID,
		},
	}

	return agentSettings
}

// AttachPersistentDisk updates the agent settings in order to add an attached persistent disk.
func (as AgentSettings) AttachPersistentDisk(diskID string, updateSetting []byte) AgentSettings {
	persistenDiskSettings := make(map[string]PersistentSettings)

	var persistentSetting PersistentSettings
	json.Unmarshal(updateSetting, &persistentSetting)

	persistenDiskSettings[diskID] = persistentSetting
	as.Disks.Persistent = persistenDiskSettings

	return as
}

// AttachPersistentDisk updates the agent settings in order to add an attached persistent disk.
func (as AgentSettings) AttachEphemeralDisk(path string) AgentSettings {
	as.Disks.Ephemeral = path

	return as
}

// ConfigureNetworks updates the agent settings with the networks settings.
func (as AgentSettings) ConfigureNetworks(networksSettings NetworksSettings) AgentSettings {
	as.Networks = networksSettings

	return as
}

// DetachPersistentDisk updates the agent settings in order to delete an attached persistent disk.
func (as AgentSettings) DetachPersistentDisk(diskID string) AgentSettings {
	persistenDiskSettings := as.Disks.Persistent
	delete(persistenDiskSettings, diskID)
	as.Disks.Persistent = persistenDiskSettings

	return as
}
