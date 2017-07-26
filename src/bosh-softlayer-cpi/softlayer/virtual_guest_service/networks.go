package instance

import (
	"bosh-softlayer-cpi/registry"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Networks map[string]Network

func (n Networks) Validate() error {
	var networks, vipNetworks int

	for _, network := range n {
		if err := network.Validate(); err != nil {
			return err
		}

		switch {
		case network.IsDynamic():
			networks++
		case network.IsManual():
			networks++
		case network.IsVip():
			vipNetworks++
		}
	}

	//if networks != 1 {
	//	return bosherr.Error("Exactly one Dynamic or Manual network must be defined")
	//}

	if vipNetworks > 1 {
		return bosherr.Error("Only one VIP network is allowed")
	}

	return nil
}

func (n Networks) Network() Network {
	for _, net := range n {
		if !net.IsVip() {
			// There can only be 1 dynamic or manual network
			return net
		}
	}

	return Network{}
}

func (n Networks) VipNetwork() Network {
	for _, net := range n {
		if net.IsVip() {
			// There can only be 1 vip network
			return net
		}
	}

	return Network{}
}

func (n Networks) DNS() []string {
	network := n.Network()

	return network.DNS
}

func (n Networks) AsRegistryNetworks() registry.NetworksSettings {
	networksSettings := registry.NetworksSettings{}

	for netName, network := range n {
		networksSettings[netName] = registry.NetworkSettings{
			Type:    network.Type,
			IP:      network.IP,
			Mac:     network.MAC,
			Gateway: network.Gateway,
			Netmask: network.Netmask,
			DNS:     network.DNS,
			Default: network.Default,
			Alias:   network.Alias,
			Routes:  network.Routes,
		}
	}

	return networksSettings
}
