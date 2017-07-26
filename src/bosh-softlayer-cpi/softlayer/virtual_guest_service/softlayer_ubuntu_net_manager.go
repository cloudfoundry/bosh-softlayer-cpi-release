package instance

import (
	"errors"
	"fmt"

	"bosh-softlayer-cpi/registry"
	datatypes "github.com/softlayer/softlayer-go/datatypes"
)

func SoftlayerPrivateRoutes(gateway string) registry.Routes {
	return []registry.Route{
		{Destination: "10.0.0.0", NetMask: "255.0.0.0", Gateway: gateway},
		{Destination: "161.26.0.0", NetMask: "255.255.0.0", Gateway: gateway},
	}
}

type Softlayer_Ubuntu_Net struct {
	LinkNamer LinkNamer
}

func (u *Softlayer_Ubuntu_Net) NormalizeNetworkDefinitions(networks Networks, componentByNetwork map[string]datatypes.Virtual_Guest_Network_Component) (Networks, error) {
	normalized := Networks{}

	for name, nw := range networks {
		switch nw.Type {
		case "dynamic":
			c := componentByNetwork[name]
			nw.IP = *c.PrimaryIpAddress
			nw.MAC = *c.MacAddress
			normalized[name] = nw
		case "manual", "":
			nw.Type = "manual"
			normalized[name] = nw
		default:
			return nil, fmt.Errorf("unexpected network type: %s", nw.Type)
		}
	}

	return normalized, nil
}

func (u *Softlayer_Ubuntu_Net) FinalizedNetworkDefinitions(networkComponents datatypes.Virtual_Guest, networks Networks, componentByNetwork map[string]datatypes.Virtual_Guest_Network_Component) (Networks, error) {
	finalized := Networks{}
	for name, nw := range networks {
		component, ok := componentByNetwork[name]
		if !ok {
			return networks, fmt.Errorf("network not found: %q", name)
		}

		for _, ipAddressBinding := range component.IpAddressBindings {
			if *ipAddressBinding.Type == "PRIMARY" {
				networkComponentIpAddress := ipAddressBinding.IpAddress
				if networkComponentIpAddress.IpAddress != nil && *networkComponentIpAddress.IpAddress == nw.IP {
					nw.Netmask = *networkComponentIpAddress.Subnet.Netmask
					nw.Gateway = *networkComponentIpAddress.Subnet.Gateway
					nw.MAC = *component.MacAddress
					if component.NetworkVlan.Id == networkComponents.PrimaryBackendNetworkComponent.NetworkVlan.Id {
						nw.Routes = SoftlayerPrivateRoutes(*networkComponentIpAddress.Subnet.Gateway)
						nw.Gateway = ""
					}
				}
			}
		}

		var alias string
		var err error

		alias = fmt.Sprintf("%s%d", *component.Name, *component.Port)
		if nw.Type != "dynamic" {
			alias, err = u.LinkNamer.Name(alias, name)
			if err != nil {
				return networks, fmt.Errorf("Linking network with name `%s`: `%s`", name, err.Error())
			}
			nw.Gateway = ""
		}
		nw.Alias = alias

		finalized[name] = nw
	}

	return finalized, nil
}

func (u *Softlayer_Ubuntu_Net) NormalizeDynamics(networkComponents datatypes.Virtual_Guest, networks Networks) (Networks, error) {
	var privateDynamic, publicDynamic *Network

	for _, nw := range networks {
		if nw.Type != "dynamic" {
			continue
		}

		if nw.CloudProperties.VlanID == *(networkComponents.PrimaryBackendNetworkComponent.NetworkVlan.Id) {
			if privateDynamic != nil {
				return nil, errors.New("multiple private dynamic networks are not supported")
			}
			privateDynamic = &nw
		}

		if nw.CloudProperties.VlanID == *(networkComponents.PrimaryNetworkComponent.NetworkVlan.Id) {
			if publicDynamic != nil {
				return nil, errors.New("multiple public dynamic networks are not supported")
			}
			publicDynamic = &nw
		}
	}

	if privateDynamic == nil {
		networks["generated-private"] = Network{
			Type: "dynamic",
			IP:   *(networkComponents.PrimaryBackendNetworkComponent.PrimaryIpAddress),
			CloudProperties: NetworkCloudProperties{
				VlanID:              *(networkComponents.PrimaryBackendNetworkComponent.NetworkVlan.Id),
				SourcePolicyRouting: true,
			},
		}
	}

	if publicDynamic == nil && networkComponents.PrimaryNetworkComponent.NetworkVlan.Id != nil {
		networks["generated-public"] = Network{
			Type: "dynamic",
			IP:   *(networkComponents.PrimaryNetworkComponent.PrimaryIpAddress),
			CloudProperties: NetworkCloudProperties{
				VlanID:              *(networkComponents.PrimaryNetworkComponent.NetworkVlan.Id),
				SourcePolicyRouting: true,
			},
		}
	}

	return networks, nil
}

func (u *Softlayer_Ubuntu_Net) ComponentByNetworkName(components datatypes.Virtual_Guest, networks Networks) (map[string]datatypes.Virtual_Guest_Network_Component, error) {
	componentByNetwork := map[string]datatypes.Virtual_Guest_Network_Component{}

	for name, network := range networks {
		switch network.CloudProperties.VlanID {
		case *components.PrimaryBackendNetworkComponent.NetworkVlan.Id:
			componentByNetwork[name] = *components.PrimaryBackendNetworkComponent
		case *components.PrimaryNetworkComponent.NetworkVlan.Id:
			componentByNetwork[name] = *components.PrimaryNetworkComponent
		default:
			return nil, fmt.Errorf("Network %q specified a vlan id '%d' that is not associated with this virtual guest", name, network.CloudProperties.VlanID)
		}
	}

	return componentByNetwork, nil
}

//go:generate counterfeiter -o fakes/fake_link_namer.go --fake-name FakeLinkNamer . LinkNamer
type LinkNamer interface {
	Name(interfaceName, networkName string) (string, error)
}

type indexedNamer struct {
	indices map[string]int
}

func NewIndexedNamer(networks Networks) LinkNamer {
	indices := map[string]int{}

	index := 0
	for name := range networks {
		indices[name] = index
		index++
	}

	return &indexedNamer{
		indices: indices,
	}
}

func (l *indexedNamer) Name(interfaceName, networkName string) (string, error) {
	idx, ok := l.indices[networkName]
	if !ok {
		return "", fmt.Errorf("Network name not found: %q", networkName)
	}

	return fmt.Sprintf("%s:%d", interfaceName, idx), nil
}
