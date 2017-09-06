package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"
)

var _ = Describe("Networks", func() {
	var ()

	BeforeEach(func() {
	})

	Describe("Call AsInstanceServiceNetworks", func() {
		It("Generate InstanceServiceNetworks with single setting", func() {
			networks := Networks{
				"fake-network-name": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanIds:             []int{42345678},
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}
			networks.AsInstanceServiceNetworks()
			ret := networks.HasManualNetwork()
			Expect(ret).To(BeFalse())
		})

		It("Generate InstanceServiceNetworks with two settings", func() {
			networks := Networks{
				"fake-network-name1": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanIds:             []int{42345678},
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

						VlanIds:             []int{42345678},
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}
			networks.AsInstanceServiceNetworks()
			ret := networks.HasManualNetwork()
			Expect(ret).To(BeTrue())
		})

		It("Generate InstanceServiceNetworks with two vlans", func() {
			networks := Networks{
				"fake-network-name": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanIds:             []int{42345678, 42345680},
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}
			networks.AsInstanceServiceNetworks()
			ret := networks.HasManualNetwork()
			Expect(ret).To(BeFalse())
		})

		It("Generate InstanceServiceNetworks with two subnets", func() {
			networks := Networks{
				"fake-network-name": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanIds:             []int{42345678, 42345680},
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}
			networks.AsInstanceServiceNetworks()
			ret := networks.HasManualNetwork()
			Expect(ret).To(BeFalse())
		})
	})
})
