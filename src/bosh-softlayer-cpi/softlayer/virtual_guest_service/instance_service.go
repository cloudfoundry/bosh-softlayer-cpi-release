package instance

import (
	"github.com/softlayer/softlayer-go/datatypes"

	"bosh-softlayer-cpi/registry"
)

//go:generate counterfeiter -o fakes/fake_Instance_Service.go . Service
type Service interface {
	AttachDisk(id int, diskID int) ([]byte, error)
	AttachedDisks(id int) ([]string, error)
	AttachEphemeralDisk(id int, diskSize int) error
	Create(virtualGuest *datatypes.Virtual_Guest, enableVps bool, stemcellID int, sshKeys []int, userData *registry.SoftlayerUserData) (int, error)
	UpgradeInstance(id int, cpu int, memory int, network int, privateCPU bool, dedicatedHost bool) error
	ConfigureNetworks(id int, networks Networks) (Networks, error)
	CleanUp(id int) error
	CreateSshKey(label string, key string, fingerPrint string) (int, error)
	Delete(id int, enableVps bool) error
	DetachDisk(id int, diskID int) error
	DeleteSshKey(id int) error
	Edit(id int, instance *datatypes.Virtual_Guest) error
	Find(id int) (*datatypes.Virtual_Guest, error)
	FindByPrimaryBackendIp(ip string) (*datatypes.Virtual_Guest, error)
	FindByPrimaryIp(ip string) (*datatypes.Virtual_Guest, error)
	GetVlan(id int, mask string) (*datatypes.Network_Vlan, error)
	GetSubnet(id int, mask string) (*datatypes.Network_Subnet, error)
	Reboot(id int) error
	ReloadOS(id int, stemcellID int, sshKeyIds []int, hostname string, domain string, userData *registry.SoftlayerUserData) error
	SetMetadata(id int, vmMetadata Metadata) error
	UpdateInstanceUserData(id int, userData *string) error
}

type Metadata map[string]interface{}

type DavConfig map[string]interface{}
