package vm

type Networks map[string]Network

type Network struct {
	Type string `json:"type"`

	IP      string `json:"ip,omitempty"`
	Netmask string `json:"netmask,omitempty"`
	Gateway string `json:"gateway,omitempty"`

	DNS     []string `json:"dns,omitempty"`
	Default []string `json:"default,omitempty"`

	Preconfigured bool `json:"preconfigured,omitempty"`

	MAC string `json:"mac,omitempty"`

	CloudProperties map[string]interface{} `json:"cloud_properties,omitempty"`
}

func (ns Networks) First() Network {
	for _, net := range ns {
		return net
	}

	return Network{}
}

func (n Network) IsDynamic() bool { return n.Type == "dynamic" }

func (n Network) AppendDNS(dns string) Network {
	if len(dns) > 0 {
		n.DNS = append(n.DNS, dns)
		return n
	}
	return n
}
