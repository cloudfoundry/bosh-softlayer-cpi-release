package action

import (
	"bosh-softlayer-cpi/registry"

	"bosh-softlayer-cpi/softlayer/virtual_guest_service"
	"fmt"
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

func (ns Networks) AsInstanceServiceNetworks() instance.Networks {
	networks := instance.Networks{}

	for netName, network := range ns {
		netSlim := instance.Network{
			Type:    network.Type,
			IP:      network.IP,
			Gateway: network.Gateway,
			Netmask: network.Netmask,
			DNS:     network.DNS,
		}

		// set one by one
		parseCloudProperties(networks, netName, netSlim, network.CloudProperties)
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

func parseCloudProperties(networks instance.Networks, netName string, network instance.Network, cloudProperties NetworkCloudProperties) {
	if len(cloudProperties.SubnetIds) > 0 {
		for index, subnetId := range cloudProperties.SubnetIds {
			var newNetName string
			var cloudProps instance.NetworkCloudProperties
			cloudProps = instance.NetworkCloudProperties{
				SubnetID: subnetId,
			}

			if index > 0 {
				newNetName = fmt.Sprintf("%s_%d", netName, index)
				network.CloudProperties = cloudProps
				networks[newNetName] = network
			} else {
				newNetName = netName
				cloudProps.SourcePolicyRouting = cloudProperties.SourcePolicyRouting
				network.CloudProperties = cloudProps
				networks[newNetName] = network
			}
		}
	} else if len(cloudProperties.VlanIds) > 0 {
		{
			for index, vlanId := range cloudProperties.VlanIds {
				var newNetName string
				var cloudProps instance.NetworkCloudProperties
				cloudProps = instance.NetworkCloudProperties{
					VlanID: vlanId,
				}

				if index > 0 {
					newNetName = fmt.Sprintf("%s_%d", netName, index)
					network.CloudProperties = cloudProps
					networks[newNetName] = network
				} else {
					newNetName = netName
					cloudProps.SourcePolicyRouting = cloudProperties.SourcePolicyRouting
					network.CloudProperties = cloudProps
					networks[newNetName] = network
				}
			}
		}
	}
}
