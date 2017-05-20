package action

import (
	"bosh-softlayer-cpi/registry"

	"bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

type Networks map[string]Network

type Network struct {
	Type            string                 `json:"type,omitempty"`
	IP              string                 `json:"ip,omitempty"`
	Gateway         string                 `json:"gateway,omitempty"`
	Netmask         string                 `json:"netmask,omitempty"`
	DNS             []string               `json:"dns,omitempty"`
	DHCP            bool                   `json:"use_dhcp,omitempty"`
	Default         []string               `json:"default,omitempty"`
	MAC             string                 `json:"mac,omitempty"`
	Alias           string                 `json:"alias,omitempty"`
	Routes          registry.Routes        `json:"routes,omitempty"`
	CloudProperties NetworkCloudProperties `json:"cloud_properties,omitempty"`
}

func (ns Networks) AsInstanceServiceNetworks() instance.Networks {
	networks := instance.Networks{}

	for netName, network := range ns {
		networks[netName] = instance.Network{
			Type:    network.Type,
			IP:      network.IP,
			Gateway: network.Gateway,
			Netmask: network.Netmask,
			DNS:     network.DNS,
			Default: network.Default,
			CloudProperties: instance.NetworkCloudProperties{
				VlanID:              network.CloudProperties.VlanID,
				SourcePolicyRouting: network.CloudProperties.SourcePolicyRouting,
			},
		}
	}

	return networks
}
