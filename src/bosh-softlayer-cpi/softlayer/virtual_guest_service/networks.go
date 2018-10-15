package instance

import (
	"bosh-softlayer-cpi/registry"
	"sort"
)

type Networks map[string]Network

func (n Networks) Validate() error {
	var networks int
	var vipNetworks int
	for _, network := range n {
		if err := network.validate(); err != nil {
			return err
		}

		switch {
		case network.isDynamic():
			networks++
		case network.isManual():
			networks++
		case network.isVip():
			vipNetworks++
		}
	}

	// if networks != 1 {
	// 	return bosherr.Error("Exactly one Dynamic or Manual network must be defined")
	// }
	//
	// // Network type 'vip' not supported currently()
	// if vipNetworks > 1 {
	// 	return bosherr.Error("Only one VIP network is allowed")
	// }

	return nil
}

func (n Networks) Network() Network {
	var keys []string
	for key, _ := range n {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if !n[key].isVip() {
			// There can only be 1 dynamic or manual network
			return n[key]
		}
	}

	return Network{}
}

// func (n Networks) VipNetwork() Network {
// 	for _, net := range n {
// 		if net.IsVip() {
// 			// There can only be 1 vip network
// 			return net
// 		}
// 	}
//
// 	return Network{}
// }

func (n Networks) DNS() []string {
	network := n.Network()

	return network.DNS
}

func (n Networks) AsRegistryNetworks() registry.NetworksSettings {
	networksSettings := registry.NetworksSettings{}

	for netName, network := range n {
		networksSettings[netName] = registry.NetworkSettings{
			Type:          network.Type,
			IP:            network.IP,
			Mac:           network.MAC,
			Gateway:       network.Gateway,
			Netmask:       network.Netmask,
			DNS:           network.DNS,
			Default:       network.Default,
			Alias:         network.Alias,
			Routes:        network.Routes,
			Preconfigured: len(network.DNS) != 0,
		}
	}

	return networksSettings
}
