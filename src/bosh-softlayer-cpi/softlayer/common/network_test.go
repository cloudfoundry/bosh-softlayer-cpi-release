package common_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
)

var _ = Describe("Network", func() {
	var (
		networks, emptyNetworks                                 Networks
		dynamicNetwork, emptyNetwork, dnsNetwork                Network
		expectedNetwork                                         Network
		networkWithDefaultGateway, networkWithOutDefaultGateway Network
	)

	Describe("#First", func() {
		BeforeEach(func() {
			networks = map[string]Network{
				"fake-network0": Network{
					Type:    "fake-type",
					IP:      "fake-IP",
					Netmask: "fake-Netmask",
					Gateway: "fake-Gateway",
					DNS: []string{
						"fake-dns0",
						"fake-dns1",
					},
					Default:         []string{},
					Preconfigured:   true,
					CloudProperties: map[string]interface{}{},
				},
			}

			emptyNetworks = map[string]Network{}

			expectedNetwork = Network{
				Type:    "fake-type",
				IP:      "fake-IP",
				Netmask: "fake-Netmask",
				Gateway: "fake-Gateway",
				DNS: []string{
					"fake-dns0",
					"fake-dns1",
				},
				Default:         []string{},
				Preconfigured:   true,
				CloudProperties: map[string]interface{}{},
			}
		})

		It("return first network in networks", func() {
			fakeNetwork := networks.First()
			Expect(fakeNetwork).ToNot(Equal(Networks{}))
			Expect(fakeNetwork).To(Equal(expectedNetwork))
		})

		It("return empty network in empty networks", func() {
			fakeNetwork := emptyNetworks.First()
			Expect(fakeNetwork).To(Equal(Network{}))
		})
	})

	Describe("#IsDynamic", func() {
		BeforeEach(func() {
			dynamicNetwork = Network{
				Type:    "dynamic",
				IP:      "fake-IP",
				Netmask: "fake-Netmask",
				Gateway: "fake-Gateway",
				DNS: []string{
					"fake-dns0",
					"fake-dns1",
				},
				Default:         []string{},
				Preconfigured:   true,
				CloudProperties: map[string]interface{}{},
			}

			emptyNetwork = Network{}
		})

		It("return true for a dynamic network", func() {
			result := dynamicNetwork.IsDynamic()
			Expect(result).To(BeTrue())
		})

		It("return false for an empty network", func() {
			result := emptyNetwork.IsDynamic()
			Expect(result).To(BeFalse())
		})
	})

	Describe("#HasDefaultGateway", func() {
		BeforeEach(func() {
			networkWithDefaultGateway = Network{
				Type:    "dynamic",
				IP:      "fake-IP",
				Netmask: "fake-Netmask",
				Gateway: "fake-Gateway",
				DNS: []string{
					"fake-dns0",
					"fake-dns1",
				},
				Default:         []string{"dns", "gateway"},
				Preconfigured:   true,
				CloudProperties: map[string]interface{}{},
			}

			networkWithOutDefaultGateway = Network{
				Type:    "dynamic",
				IP:      "fake-IP",
				Netmask: "fake-Netmask",
				Gateway: "fake-Gateway",
				DNS: []string{
					"fake-dns0",
					"fake-dns1",
				},
				Default:         []string{},
				Preconfigured:   true,
				CloudProperties: map[string]interface{}{},
			}

			emptyNetwork = Network{}
		})

		It("return true for a network with default gateway", func() {
			result := networkWithDefaultGateway.HasDefaultGateway()
			Expect(result).To(BeTrue())
		})

		It("return false for a network without default gateway", func() {
			result := networkWithOutDefaultGateway.HasDefaultGateway()
			Expect(result).To(BeFalse())
		})

		It("return false for an empty network", func() {
			result := emptyNetwork.HasDefaultGateway()
			Expect(result).To(BeFalse())
		})
	})

	Describe("#AppendDNS", func() {
		BeforeEach(func() {
			dnsNetwork = Network{
				Type:    "fake-type",
				IP:      "fake-IP",
				Netmask: "fake-Netmask",
				Gateway: "fake-Gateway",
				DNS: []string{
					"fake-dns0",
					"fake-dns1",
				},
				Default:         []string{},
				Preconfigured:   true,
				CloudProperties: map[string]interface{}{},
			}

			expectedNetwork = Network{
				Type:    "fake-type",
				IP:      "fake-IP",
				Netmask: "fake-Netmask",
				Gateway: "fake-Gateway",
				DNS: []string{
					"fake-dns0",
					"fake-dns1",
					"fake-dns2",
				},
				Default:         []string{},
				Preconfigured:   true,
				CloudProperties: map[string]interface{}{},
			}
		})

		It("return network with new DNS appended", func() {
			dns2Network := dnsNetwork.AppendDNS("fake-dns2")
			Expect(dns2Network).To(Equal(expectedNetwork))
		})

		It("returns network with no DNS appended", func() {
			network2 := dnsNetwork.AppendDNS("")
			Expect(network2).To(Equal(dnsNetwork))
		})
	})
})
