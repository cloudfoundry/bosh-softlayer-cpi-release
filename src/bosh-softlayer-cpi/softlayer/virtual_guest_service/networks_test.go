package instance_test

import (
	//"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bosh-softlayer-cpi/registry"
	. "bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

var _ = Describe("Networks", func() {
	var (
		networks Networks
	)

	Describe("Validate", func() {
		It("Validate single dynamic network successfully", func() {
			networks = Networks{
				"fake-network-name": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}

			err := networks.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Validate single manual network successfully", func() {
			networks = Networks{
				"fake-network-name": Network{
					Type:    "manual",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}

			err := networks.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when validate single vip network", func() {
			networks = Networks{
				"fake-network-name": Network{
					Type:    "vip",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}

			err := networks.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Network type 'vip' not supported"))
		})
	})

	Describe("Network", func() {
		It("Return first network when have one network", func() {
			networks = Networks{
				"fake-network-name1": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}

			network := networks.Network()
			Expect(network.Type).To(Equal("dynamic"))
		})

		It("Return first network when have more than one network", func() {
			networks = Networks{
				"fake-network-name1": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
				"fake-network-name2": Network{
					Type:    "manual",
					IP:      "12.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}

			network := networks.Network()
			Expect(network.Type).To(Equal("dynamic"))
		})

		It("Return empty network when have no network", func() {
			networks = Networks{}

			network := networks.Network()
			Expect(network).To(Equal(Network{}))
		})
	})

	Describe("DNS", func() {
		It("Return the dns of first network when have one network", func() {
			networks = Networks{
				"fake-network-name1": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns1"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}

			dns := networks.DNS()
			Expect(dns).To(Equal([]string{"fake-network-dns1"}))
		})

		It("Return the dns of first network when have more than one network", func() {
			networks = Networks{
				"fake-network-name1": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns1"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
				"fake-network-name2": Network{
					Type:    "dynamic",
					IP:      "12.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns2"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}

			network := networks.Network()
			Expect(network.Type).To(Equal("dynamic"))
		})

		It("Return empty network when have no network", func() {
			networks = Networks{}

			dns := networks.DNS()
			Expect(dns).To(BeNil())
		})
	})

	Describe("AsRegistryNetworks", func() {
		It("Return networksSettings when have one network", func() {
			networks = Networks{
				"fake-network-name1": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns1"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
				"fake-network-name2": Network{
					Type:    "manual",
					IP:      "12.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns2"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}

			expectNetworksSettings := registry.NetworksSettings{
				"fake-network-name1": registry.NetworkSettings{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns1"},
					Default: []string{"fake-network-default"},
				},
				"fake-network-name2": registry.NetworkSettings{
					Type:    "manual",
					IP:      "12.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns2"},
					Default: []string{"fake-network-default"},
				},
			}

			networksSettings := networks.AsRegistryNetworks()
			Expect(networksSettings).To(Equal(expectNetworksSettings))
		})
	})
})
