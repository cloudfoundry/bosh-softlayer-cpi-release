package action

import (
	"bosh-softlayer-cpi/registry"

	"bosh-softlayer-cpi/softlayer/virtual_guest_service"
	"fmt"
	"github.com/softlayer/softlayer-go/datatypes"
)

const (
	NetworkTypeManual string = "manual"
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

func (ns Networks) AsInstanceServiceNetworks(publicNetworkVlan *datatypes.Network_Vlan) instance.Networks {
	networks := instance.Networks{}

	for netName, network := range ns {
		parseCloudProperties(networks, netName, network, publicNetworkVlan)
	}

	return networks
}

func (ns Networks) HasManualNetwork() bool {
	for _, network := range ns {
		if network.IsManual() {
			return true
		}
	}

	return false
}

func (n Network) IsManual() bool {
	return n.Type == NetworkTypeManual
}

func parseCloudProperties(networks instance.Networks, netName string, network Network, publicNetworkVlan *datatypes.Network_Vlan) {
	if len(network.CloudProperties.SubnetIds) > 0 {
		for index, subnetId := range network.CloudProperties.SubnetIds {
			var newNetName string
			if publicNetworkVlan.Id != nil && *publicNetworkVlan.PrimarySubnetId == subnetId && network.Type != "manual" {
				newNetName = netName
				networks[newNetName] = instance.Network{
					Type:    network.Type,
					IP:      network.IP,
					Gateway: network.Gateway,
					Netmask: network.Netmask,
					DNS:     network.DNS,
					CloudProperties: instance.NetworkCloudProperties{
						SubnetID:            subnetId,
						SourcePolicyRouting: network.CloudProperties.SourcePolicyRouting,
					},
					Default: network.Default,
				}
			} else {
				if index != 0 {
					newNetName = fmt.Sprintf("%s_%d", netName, index)
				}else {
					newNetName = netName
				}
				networks[newNetName] = instance.Network{
					Type:    network.Type,
					IP:      network.IP,
					Gateway: network.Gateway,
					Netmask: network.Netmask,
					DNS:     network.DNS,
					CloudProperties: instance.NetworkCloudProperties{
						SubnetID: subnetId,
					},
				}
			}
		}
	} else if len(network.CloudProperties.VlanIds) > 0 {
		for index, vlanId := range network.CloudProperties.VlanIds {
			var newNetName string
			if publicNetworkVlan.Id != nil && *publicNetworkVlan.Id == vlanId && network.Type != "manual" {
				newNetName = netName
				networks[newNetName] = instance.Network{
					Type:    network.Type,
					IP:      network.IP,
					Gateway: network.Gateway,
					Netmask: network.Netmask,
					DNS:     network.DNS,
					CloudProperties: instance.NetworkCloudProperties{
						VlanID:              vlanId,
						SourcePolicyRouting: network.CloudProperties.SourcePolicyRouting,
					},
					Default: network.Default,
				}
			} else {
				if index != 0 {
					newNetName = fmt.Sprintf("%s_%d", netName, index)
				}else {
					newNetName = netName
				}
				networks[newNetName] = instance.Network{
					Type:    network.Type,
					IP:      network.IP,
					Gateway: network.Gateway,
					Netmask: network.Netmask,
					DNS:     network.DNS,
					CloudProperties: instance.NetworkCloudProperties{
						VlanID: vlanId,
					},
				}
			}
		}
	}
}
