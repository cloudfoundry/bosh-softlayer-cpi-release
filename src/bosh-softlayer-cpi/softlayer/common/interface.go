package common

import (
	bslcdisk "bosh-softlayer-cpi/softlayer/disk"
	bslcstem "bosh-softlayer-cpi/softlayer/stemcell"
	"encoding/json"
	sldatatypes "github.com/maximilien/softlayer-go/data_types"
	"reflect"
	"strings"
	"unicode"
)

type Environment map[string]interface{}

type Mount struct {
	PartitionPath string
	MountPoint    string
}

type FeatureOptions struct {
	DisableOsReload                  bool   `json:"disableOsReload"`
	EnablePool                       bool   `json:"enablePool"`
	ApiEndpoint                      string `json:"apiEndpoint"`
	ApiWaitTime                      int    `json:"apiWaitTime"`
	ApiRetryCount                    int    `json:"apiRetryCount"`
	CreateISCSIVolumeTimeout         int    `json:"createIscsiVolumeTimeout"`
	CreateISCSIVolumePollingInterval int    `json:"createIscsiVolumePollingInterval"`
}

type VMCloudProperties struct {
	Hostname                 string                               `json:"hostname,omitempty"`
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

	BoshIp            string `json:"bosh_ip,omitempty"`
	DeployedByBoshCLI bool   `json:"deployedByBoshcli,omitempty"`

	Baremetal             bool   `json:"baremetal,omitempty"`
	BaremetalStemcell     string `json:"bm_stemcell,omitempty"`
	BaremetalNetbootImage string `json:"bm_netboot_image,omitempty"`

	DisableOsReload bool `json:"disableOsReload,omitempty"`
}

func (vmprop *VMCloudProperties) UnmarshalJSON(data []byte) error {
	type vmCloudProperties VMCloudProperties
	err := json.Unmarshal(data, (*vmCloudProperties)(vmprop))
	if err != nil {
		return err
	}
	var oriProps map[string]interface{}
	err = json.Unmarshal(data, &oriProps)
	if err != nil {
		return err
	}
	converted, err := ConvertKeysCamelized(oriProps)
	if err != nil {
		return err
	}
	j, err := json.Marshal(converted)
	if err != nil {
		return err
	}
	err = json.Unmarshal(j, (*vmCloudProperties)(vmprop))
	if err != nil {
		return err
	}
	return nil
}

func ConvertKeysCamelized(ori interface{}) (interface{}, error) {
	rv := reflect.ValueOf(ori)
	switch rv.Kind() {
	case reflect.Slice:
		ret := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			v, err := ConvertKeysCamelized(rv.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			ret[i] = v
		}
		return ret, nil
	case reflect.Map:
		ret := make(map[string]interface{})
		for _, k := range rv.MapKeys() {
			new_k := camelize(k.Interface().(string))
			v := rv.MapIndex(k).Interface()
			new_v, err := ConvertKeysCamelized(v)
			if err != nil {
				return nil, err
			}
			ret[new_k] = new_v
		}
		return ret, nil
	default:
		return rv.Interface(), nil

	}
}

func camelize(in string) string {
	var res string
	words := strings.Split(in, "_")
	for _, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			res += string(runes)
		} else {
			res += word
		}
	}
	return res
}

type AllowedHostCredential struct {
	Iqn      string `json:"iqn"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type VMMetadata map[string]interface{}

const EtcIscsidConfTemplate = `# Generated by bosh-agent
node.startup = automatic
node.session.auth.authmethod = CHAP
node.session.auth.username = {{.Username}}
node.session.auth.password = {{.Password}}
discovery.sendtargets.auth.authmethod = CHAP
discovery.sendtargets.auth.username = {{.Username}}
discovery.sendtargets.auth.password = {{.Password}}
node.session.timeo.replacement_timeout = 120
node.conn[0].timeo.login_timeout = 15
node.conn[0].timeo.logout_timeout = 15
node.conn[0].timeo.noop_out_interval = 10
node.conn[0].timeo.noop_out_timeout = 15
node.session.iscsi.InitialR2T = No
node.session.iscsi.ImmediateData = Yes
node.session.iscsi.FirstBurstLength = 262144
node.session.iscsi.MaxBurstLength = 16776192
node.conn[0].iscsi.MaxRecvDataSegmentLength = 65536
`

const (
	SOFTLAYER_HARDWARE_LOG_TAG   = "SoftLayerHardware"
	SOFTLAYER_VM_FINDER_LOG_TAG  = "SoftLayerVMFinder"
	SOFTLAYER_VM_OS_RELOAD_TAG   = "OSReload"
	SOFTLAYER_VM_LOG_TAG         = "SoftLayerVM"
	ROOT_USER_NAME               = "root"
	SOFTLAYER_VM_CREATOR_LOG_TAG = "SoftLayerVMCreator"
)

//go:generate counterfeiter -o fakes/fake_vm.go . VM
type VM interface {
	AttachDisk(bslcdisk.Disk) error

	ConfigureNetworks(Networks) error

	//dedicated for setup network by modify /etc/network/interfaces
	ConfigureNetworks2(Networks) error

	DetachDisk(bslcdisk.Disk) error
	Delete(agentId string) error

	GetDataCenterId() int
	GetPrimaryIP() string
	GetPrimaryBackendIP() string
	GetRootPassword() string
	GetFullyQualifiedDomainName() string

	ID() int

	Reboot() error
	ReloadOS(bslcstem.Stemcell) error
	ReloadOSForBaremetal(string, string) error

	SetMetadata(VMMetadata) error
	SetVcapPassword(string) error
	SetAgentEnvService(AgentEnvService) error

	UpdateAgentEnv(AgentEnv) error
	DeleteAgentEnv() error
}

//go:generate counterfeiter -o fakes/fake_vm_creator.go . VMCreator
type VMCreator interface {
	Create(string, bslcstem.Stemcell, VMCloudProperties, Networks, Environment) (VM, error)
	GetAgentOptions() AgentOptions
}

//go:generate counterfeiter -o fakes/fake_vm_deleter.go . VMDeleter
type VMDeleter interface {
	Delete(cid int) error
}

//go:generate counterfeiter -o fakes/fake_vm_finder.go . VMFinder
type VMFinder interface {
	Find(int) (VM, bool, error)
}
