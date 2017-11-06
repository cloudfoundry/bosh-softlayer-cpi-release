package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("Networks", func() {
	var (
		publicVlan *datatypes.Network_Vlan
	)

	BeforeEach(func() {
		publicVlan = &datatypes.Network_Vlan{
			Id:           sl.Int(42345680),
			NetworkSpace: sl.String("PUBLIC"),
		}
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
					},
				},
			}
			networks.AsInstanceServiceNetworks(
				&datatypes.Network_Vlan{})
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
					},
				},
			}
			networks.AsInstanceServiceNetworks(&datatypes.Network_Vlan{})
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
					},
				},
			}
			networks.AsInstanceServiceNetworks(publicVlan)
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
					},
				},
			}
			networks.AsInstanceServiceNetworks(publicVlan)
			ret := networks.HasManualNetwork()
			Expect(ret).To(BeFalse())
		})
	})
})
