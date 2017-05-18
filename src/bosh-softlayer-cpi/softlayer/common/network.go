package common

import "encoding/json"

//import "encoding/json"

type Networks map[string]Network
type CloudProperties map[string]interface{}

type Network struct {
	Type string `json:"type"`

	IP      string `json:"ip,omitempty"`
	Netmask string `json:"netmask,omitempty"`
	Gateway string `json:"gateway,omitempty"`

	DNS     []string `json:"dns,omitempty"`
	Default []string `json:"default,omitempty"`

	Preconfigured bool `json:"preconfigured,omitempty"`

	MAC string `json:"mac,omitempty"`

	CloudProperties CloudProperties `json:"cloud_properties,omitempty"`
}

func (cloudProp *CloudProperties) UnmarshalJSON(data []byte) error {
	//*cloudProp = CloudProperties{"s": "df"}
	//return nil
	type cloudProperties CloudProperties
	var oriProps map[string]interface{}
	err := json.Unmarshal(data, &oriProps)
	if err != nil {
		return err
	}
	converted, err := ConvertKeysCamelized(oriProps)
	if err != nil {
		return err
	}
	j, err := json.Marshal(converted)
	if err != nil {
		return err
	}

	err = json.Unmarshal(j, (*cloudProperties)(cloudProp))
	if err != nil {
		return err
	}
	return nil
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

func (n Network) IsDynamic() bool { return n.Type == "dynamic" }

func (n Network) AppendDNS(dns string) Network {
	if len(dns) > 0 {
		n.DNS = append(n.DNS, dns)
		return n
	}
	return n
}
