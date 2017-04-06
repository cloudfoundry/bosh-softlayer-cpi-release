package common

type Networks map[string]Network

type NetworkCloudProperties struct {
	VlanID              int  `json:"vlanId"`
	SourcePolicyRouting bool `json:"source_policy_routing,omitempty"`
}

type Network struct {
	Type string `json:"type"`

	IP      string `json:"ip,omitempty"`
	Netmask string `json:"netmask,omitempty"`
	Gateway string `json:"gateway,omitempty"`

	DNS     []string `json:"dns,omitempty"`
	Default []string `json:"default,omitempty"`

	MAC string `json:"mac,omitempty"`

	Alias  string  `json:"alias,omitempty"`
	Routes []Route `json:"routes,omitempty"`

	CloudProperties NetworkCloudProperties `json:"cloud_properties,omitempty"`
}

func (ns Networks) First() Network {
	for _, net := range ns {
		return net
	}

	return Network{}
}

func (n Network) HasDefaultGateway() bool {
	for _, val := range n.Default {
		if val == "gateway" {
			return true
		}
	}
	return false
}

func (n Network) SourcePolicyRouting() bool {
	return n.CloudProperties.SourcePolicyRouting
}

func (n Network) IsDynamic() bool { return n.Type == "dynamic" }

func (n Network) AppendDNS(dns string) Network {
	if len(dns) > 0 {
		n.DNS = append(n.DNS, dns)
		return n
	}
	return n
}
