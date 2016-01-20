package action

import (
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type Networks map[string]Network

type Network struct {
	Type string `json:"type"`

	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`

	DNS     []string `json:"dns"`
	Default []string `json:"default"`

	Preconfigured bool `json:"preconfigured"`

	MAC string `json:"mac"`

	CloudProperties map[string]interface{} `json:"cloud_properties"`
}

func (ns Networks) AsVMNetworks() bslcvm.Networks {
	networks := bslcvm.Networks{}

	for netName, network := range ns {
		networks[netName] = bslcvm.Network{
			Type: network.Type,

			IP:      network.IP,
			Netmask: network.Netmask,
			Gateway: network.Gateway,

			DNS:           network.DNS,
			Default:       network.Default,
			Preconfigured: true,

			CloudProperties: network.CloudProperties,
		}
	}

	return networks
}
