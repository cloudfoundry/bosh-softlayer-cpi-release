package instance

import (
	"bosh-softlayer-cpi/registry"
	"github.com/softlayer/softlayer-go/datatypes"
)

type Service interface {
	AttachDisk(id int, diskID int) (string, string, error)
	Create(vmProps *Properties, networks Networks, registryEndpoint string) (int, error)
	ConfigureNetworks(id int, networks Networks) (Networks, error)
	CleanUp(id int)
	Delete(id int) error
	DetachDisk(id int, diskID int) error
	Find(id int) (datatypes.Virtual_Guest, bool, error)
	GetVlan(id int, mask string) (datatypes.Network_Vlan, error)
	Reboot(id int) error
	ReloadOS(id int, stemcellID int) error
	SetMetadata(id int, vmMetadata Metadata) error
}

type Metadata map[string]interface{}

type Properties struct {
	VirtualGuestTemplate datatypes.Virtual_Guest
	SecondDisk           int
	DeployedByBoshCLI    bool
	agentOption          registry.AgentOptions
}
