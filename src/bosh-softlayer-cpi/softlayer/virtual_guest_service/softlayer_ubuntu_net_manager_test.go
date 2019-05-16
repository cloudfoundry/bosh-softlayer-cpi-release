package instance_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"

	"bosh-softlayer-cpi/registry"
	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	. "bosh-softlayer-cpi/softlayer/virtual_guest_service"
	fakesVirtualGustService "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"
)

var _ = Describe("Virtual Guest Service", func() {
	var (
		cli *fakeslclient.FakeClient

		net *Softlayer_Ubuntu_Net
	)

	BeforeEach(func() {
		cli = &fakeslclient.FakeClient{}

		net = &Softlayer_Ubuntu_Net{
			LinkNamer: &fakesVirtualGustService.FakeLinkNamer{},
		}
	})

	Describe("Call SoftlayerPrivateRoutes", func() {
		var (
			gateway string
		)

		BeforeEach(func() {
			gateway = "fake-gateway"
		})

		Context("Generate registry routes", func() {
			It("Generate routes successfully", func() {
				cli.DeleteInstanceFromVPSReturns(
					nil,
				)

				expectRoutes := []registry.Route{
					{Destination: "10.0.0.0", NetMask: "255.0.0.0", Gateway: gateway},
					{Destination: "161.26.0.0", NetMask: "255.255.0.0", Gateway: gateway},
				}

				routes := SoftlayerPrivateRoutes(gateway)
				Expect(routes).To(BeEquivalentTo(expectRoutes))
			})
		})
	})

	Describe("Call FinalizedNetworkDefinitions", func() {
		var (
			networkComponents  datatypes.Virtual_Guest
			networks           Networks
			componentByNetwork map[string]datatypes.Virtual_Guest_Network_Component
		)

		BeforeEach(func() {
			networkComponents = datatypes.Virtual_Guest{
				PrimaryBackendNetworkComponent: &datatypes.Virtual_Guest_Network_Component{
					NetworkVlan: &datatypes.Network_Vlan{
						Id: sl.Int(12345678),
					},
				},
				PrimaryNetworkComponent: &datatypes.Virtual_Guest_Network_Component{
					NetworkVlan: &datatypes.Network_Vlan{
						Id: sl.Int(1234580),
					},
				},
			}
			networks = Networks{
				"fake-network1": Network{
					Type:            "dynamic",
					IP:              "fake-ip-address1",
					Netmask:         "fake-netmask",
					Gateway:         "fake-gateway",
					DNS:             []string{"fake-dns"},
					Default:         []string{""},
					MAC:             "fake-mac-address1",
					Alias:           "fake-alias",
					Routes:          registry.Routes{},
					CloudProperties: NetworkCloudProperties{},
				},
				"fake-network2": Network{
					Type:            "manual",
					IP:              "fake-ip-address2",
					Netmask:         "fake-netmask",
					Gateway:         "fake-gateway",
					DNS:             []string{"fake-dns"},
					Default:         []string{""},
					MAC:             "fake-mac-address2",
					Alias:           "fake-alias",
					Routes:          registry.Routes{},
					CloudProperties: NetworkCloudProperties{},
				},
			}

			componentByNetwork = map[string]datatypes.Virtual_Guest_Network_Component{
				"fake-network1": {
					IpAddressBindings: []datatypes.Virtual_Guest_Network_Component_IpAddress{
						{
							Type:        sl.String("PRIMARY"),
							IpAddressId: sl.Int(32345678),
							IpAddress: &datatypes.Network_Subnet_IpAddress{
								Id:        sl.Int(32345678),
								IpAddress: sl.String("fake-ip-address1"),
								Subnet: &datatypes.Network_Subnet{
									Netmask: sl.String("fake-netmask1"),
									Gateway: sl.String("fake-gateway1"),
								},
							},
						},
					},
					NetworkVlan: &datatypes.Network_Vlan{
						Id: sl.Int(12345678),
					},
					MacAddress: sl.String("fake-mac-address1"),
					Name:       sl.String("eth"),
					Port:       sl.Int(0),
				},
				"fake-network2": {
					IpAddressBindings: []datatypes.Virtual_Guest_Network_Component_IpAddress{
						{
							Type:        sl.String("PRIMARY"),
							IpAddressId: sl.Int(42345678),
							IpAddress: &datatypes.Network_Subnet_IpAddress{
								Id:        sl.Int(42345678),
								IpAddress: sl.String("fake-ip-address2"),
								Subnet: &datatypes.Network_Subnet{
									Netmask: sl.String("fake-netmask2"),
									Gateway: sl.String("fake-gateway2"),
								},
							},
						},
					},
					NetworkVlan: &datatypes.Network_Vlan{
						Id: sl.Int(1234580),
					},
					MacAddress: sl.String("fake-mac-address2"),
					Name:       sl.String("eth"),
					Port:       sl.Int(1),
				},
			}
		})

		Context("Generate final networks", func() {
			It("Generate networks successfully", func() {
				_, err := net.FinalizedNetworkDefinitions(networkComponents, networks, componentByNetwork)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Return error when network not found", func() {
				componentByNetwork = map[string]datatypes.Virtual_Guest_Network_Component{
					"fake-network3": {
						IpAddressBindings: []datatypes.Virtual_Guest_Network_Component_IpAddress{
							{
								Type:        sl.String("PRIMARY"),
								IpAddressId: sl.Int(32345678),
								IpAddress: &datatypes.Network_Subnet_IpAddress{
									Id:        sl.Int(32345678),
									IpAddress: sl.String("fake-ip-address1"),
									Subnet: &datatypes.Network_Subnet{
										Netmask: sl.String("fake-netmask1"),
										Gateway: sl.String("fake-gateway1"),
									},
								},
							},
						},
						NetworkVlan: &datatypes.Network_Vlan{
							Id: sl.Int(12345678),
						},
						MacAddress: sl.String("fake-mac-address1"),
						Name:       sl.String("eth"),
						Port:       sl.Int(0),
					},
				}
				_, err := net.FinalizedNetworkDefinitions(networkComponents, networks, componentByNetwork)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("network not found"))
			})
		})
	})

	Describe("Call NormalizeDynamics", func() {
		var (
			networkComponents datatypes.Virtual_Guest
			networks          Networks
		)

		BeforeEach(func() {
			networkComponents = datatypes.Virtual_Guest{
				PrimaryBackendNetworkComponent: &datatypes.Virtual_Guest_Network_Component{
					PrimaryIpAddress: sl.String("fake-ip-address-primary1"),
					NetworkVlan: &datatypes.Network_Vlan{
						Id: sl.Int(12345678),
					},
				},
				PrimaryNetworkComponent: &datatypes.Virtual_Guest_Network_Component{
					PrimaryIpAddress: sl.String("fake-ip-address-primary1"),
					NetworkVlan: &datatypes.Network_Vlan{
						Id: sl.Int(12345680),
					},
				},
			}
			networks = Networks{
				"fake-network1": Network{
					Type:    "dynamic",
					IP:      "fake-ip-address1",
					Netmask: "fake-netmask",
					Gateway: "fake-gateway",
					DNS:     []string{"fake-dns"},
					Default: []string{""},
					MAC:     "fake-mac-address1",
					Alias:   "fake-alias",
					Routes:  registry.Routes{},
					CloudProperties: NetworkCloudProperties{
						VlanID: 12345678,
					},
				},
				"fake-network2": Network{
					Type:    "dynamic",
					IP:      "fake-ip-address2",
					Netmask: "fake-netmask",
					Gateway: "fake-gateway",
					DNS:     []string{"fake-dns"},
					Default: []string{""},
					MAC:     "fake-mac-address2",
					Alias:   "fake-alias",
					Routes:  registry.Routes{},
					CloudProperties: NetworkCloudProperties{
						VlanID: 12345680,
					},
				},
			}
		})

		Context("Normalize dynamic networks", func() {
			It("Normalize dynamic networks successfully", func() {
				_, err := net.NormalizeDynamics(networkComponents, networks)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Normalize dynamic networks successfully when have only one private dynamic networks", func() {
				networks = Networks{
					"fake-network1": Network{
						Type:    "dynamic",
						IP:      "fake-ip-address1",
						Netmask: "fake-netmask",
						Gateway: "fake-gateway",
						DNS:     []string{"fake-dns"},
						Default: []string{""},
						MAC:     "fake-mac-address1",
						Alias:   "fake-alias",
						Routes:  registry.Routes{},
						CloudProperties: NetworkCloudProperties{
							VlanID: 12345678,
						},
					},
				}
				_, err := net.NormalizeDynamics(networkComponents, networks)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Normalize dynamic networks successfully when have only one public dynamic networks", func() {
				networks = Networks{
					"fake-network2": Network{
						Type:    "dynamic",
						IP:      "fake-ip-address2",
						Netmask: "fake-netmask",
						Gateway: "fake-gateway",
						DNS:     []string{"fake-dns"},
						Default: []string{""},
						MAC:     "fake-mac-address2",
						Alias:   "fake-alias",
						Routes:  registry.Routes{},
						CloudProperties: NetworkCloudProperties{
							VlanID: 12345680,
						},
					},
				}
				_, err := net.NormalizeDynamics(networkComponents, networks)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Return error when have multiple public dynamic networks", func() {
				networks = Networks{
					"fake-network1": Network{
						Type:    "dynamic",
						IP:      "fake-ip-address1",
						Netmask: "fake-netmask",
						Gateway: "fake-gateway",
						DNS:     []string{"fake-dns"},
						Default: []string{""},
						MAC:     "fake-mac-address1",
						Alias:   "fake-alias",
						Routes:  registry.Routes{},
						CloudProperties: NetworkCloudProperties{
							VlanID: 12345680,
						},
					},
					"fake-network2": Network{
						Type:    "dynamic",
						IP:      "fake-ip-address2",
						Netmask: "fake-netmask",
						Gateway: "fake-gateway",
						DNS:     []string{"fake-dns"},
						Default: []string{""},
						MAC:     "fake-mac-address2",
						Alias:   "fake-alias",
						Routes:  registry.Routes{},
						CloudProperties: NetworkCloudProperties{
							VlanID: 12345680,
						},
					},
				}
				_, err := net.NormalizeDynamics(networkComponents, networks)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("multiple public dynamic networks are not supported"))
			})

			It("Return error when have multiple private dynamic networks", func() {
				networks = Networks{
					"fake-network1": Network{
						Type:    "dynamic",
						IP:      "fake-ip-address1",
						Netmask: "fake-netmask",
						Gateway: "fake-gateway",
						DNS:     []string{"fake-dns"},
						Default: []string{""},
						MAC:     "fake-mac-address1",
						Alias:   "fake-alias",
						Routes:  registry.Routes{},
						CloudProperties: NetworkCloudProperties{
							VlanID: 12345678,
						},
					},
					"fake-network2": Network{
						Type:    "dynamic",
						IP:      "fake-ip-address2",
						Netmask: "fake-netmask",
						Gateway: "fake-gateway",
						DNS:     []string{"fake-dns"},
						Default: []string{""},
						MAC:     "fake-mac-address2",
						Alias:   "fake-alias",
						Routes:  registry.Routes{},
						CloudProperties: NetworkCloudProperties{
							VlanID: 12345678,
						},
					},
				}
				_, err := net.NormalizeDynamics(networkComponents, networks)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("multiple private dynamic networks are not supported"))
			})
		})
	})

	Describe("LinkNamer", func() {
		var (
			networks    Networks
			indexdNamer LinkNamer
		)
		BeforeEach(func() {
			networks = Networks{
				"fake-network1": Network{
					Type:    "dynamic",
					IP:      "fake-ip-address1",
					Netmask: "fake-netmask",
					Gateway: "fake-gateway",
					DNS:     []string{"fake-dns"},
					Default: []string{""},
					MAC:     "fake-mac-address1",
					Alias:   "fake-alias",
					Routes:  registry.Routes{},
					CloudProperties: NetworkCloudProperties{
						VlanID: 12345678,
					},
				},
				"fake-network2": Network{
					Type:    "dynamic",
					IP:      "fake-ip-address2",
					Netmask: "fake-netmask",
					Gateway: "fake-gateway",
					DNS:     []string{"fake-dns"},
					Default: []string{""},
					MAC:     "fake-mac-address2",
					Alias:   "fake-alias",
					Routes:  registry.Routes{},
					CloudProperties: NetworkCloudProperties{
						VlanID: 12345680,
					},
				},
			}

			indexdNamer = NewIndexedNamer(networks)
		})

		It("Call Name successfully successfully", func() {
			_, err := indexdNamer.Name("eth0", "fake-network2")
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when networkName not found", func() {
			_, err := indexdNamer.Name("eth0", "fake-network3")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Network name not found"))
		})
	})
})
