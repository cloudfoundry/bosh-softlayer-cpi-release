package instance

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"regexp"

	"bosh-softlayer-cpi/registry"
)

type NetworkCloudProperties struct {
	VlanID              int  `json:"vlanId"`
	SourcePolicyRouting bool `json:"source_policy_routing,omitempty"`
	Tags                Tags `json:"tags,omitempty"`
}

const maxTagLength = 63

type Tags []string

func (t Tags) Validate() error {
	if len(t) > 0 {
		pattern, _ := regexp.Compile("^[A-Za-z]+[A-Za-z0-9-]*[A-Za-z0-9]+$")
		for _, tag := range t {
			if len(tag) > maxTagLength || !pattern.MatchString(tag) {
				return bosherr.Errorf("Invalid tag '%s': does not comply with RFC1035", tag)
			}
		}
	}

	return nil
}

func (t Tags) Unique() []string {
	tagDict := make(map[string]struct{})
	for _, tag := range t {
		tagDict[tag] = struct{}{}
	}

	tagItems := make([]string, 0)
	for tag, _ := range tagDict {
		tagItems = append(tagItems, tag)
	}
	return tagItems
}

type Network struct {
	Type string `json:"type"`

	IP      string `json:"ip,omitempty"`
	Netmask string `json:"netmask,omitempty"`
	Gateway string `json:"gateway,omitempty"`

	DNS     []string `json:"dns,omitempty"`
	Default []string `json:"default,omitempty"`

	MAC string `json:"mac,omitempty"`

	Alias  string          `json:"alias,omitempty"`
	Routes registry.Routes `json:"routes,omitempty"`

	CloudProperties NetworkCloudProperties `json:"cloud_properties,omitempty"`
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

func (n Network) IsVip() bool { return n.Type == "vip" }

func (n Network) IsManual() bool { return n.Type == "" || n.Type == "manual" }

func (n Network) Validate() error {
	switch {
	case n.IsVip():
		return bosherr.Errorf("Network type '%s' not supported", n.Type)

	default:
		return nil
	}
}
