package common_test

import (
	"encoding/json"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UbuntuNetwork", func() {
	var (
		networks             Networks
		networkComponents    VirtualGuestNetworkComponents
		softlayerClient      *fakescommon.FakeSoftLayerClient
		sshClient            *fakescommon.FakeSSHClient
		softlayerFileService *fakescommon.FakeSLFileService

		ubuntu *Ubuntu
	)

	BeforeEach(func() {
		networks = Networks{
			"default": Network{
				Type:    "dynamic",
				Default: []string{"gateway"},
			},
		}
		networkComponents = VirtualGuestNetworkComponents{
			PrimaryBackendNetworkComponent: NetworkComponent{
				Name:             "eth",
				Port:             0,
				PrimaryIPAddress: "10.155.248.190",
				NetworkVLAN: NetworkVLAN{
					Name: "private vlan",
					Subnets: []Subnet{{
						NetworkIdentifier: "10.155.248.160",
						Gateway:           "10.155.248.161",
						BroadcastAddress:  "10.155.248.191",
						Netmask:           "255.255.255.224",
					}},
				},
			},
			PrimaryNetworkComponent: NetworkComponent{},
		}
		softlayerClient = &fakescommon.FakeSoftLayerClient{}
		sshClient = &fakescommon.FakeSSHClient{}
		softlayerFileService = &fakescommon.FakeSLFileService{}

		ubuntu = &Ubuntu{
			SoftLayerClient:      softlayerClient,
			SSHClient:            sshClient,
			SoftLayerFileService: softlayerFileService,
		}
	})

	Describe("ConfigureNetwork", func() {
		var expectedConfig []byte

		BeforeEach(func() {
			jsonBytes, err := json.Marshal(networkComponents)
			Expect(err).NotTo(HaveOccurred())

			softlayerClient.DoRawHttpRequestReturns(jsonBytes, 200, nil)

			intfs, err := ubuntu.GetInterfaces(networks, 999)
			Expect(err).NotTo(HaveOccurred())

			expectedConfig, err = intfs.Configuration()
			Expect(err).NotTo(HaveOccurred())
		})

		It("uploads the configuration and restarts networking", func() {
			err := ubuntu.ConfigureNetwork(networks, &fakescommon.FakeVM{})
			Expect(err).NotTo(HaveOccurred())

			Expect(softlayerFileService.UploadCallCount()).To(Equal(1))

			path, data := softlayerFileService.UploadArgsForCall(0)
			Expect(path).To(Equal("/etc/network/interfaces.bosh"))
			Expect(data).To(BeEquivalentTo(expectedConfig))

			Expect(sshClient.OutputCallCount()).To(Equal(1))
			Expect(sshClient.OutputArgsForCall(0)).To(Equal("bash -c 'ifdown -a && mv /etc/network/interfaces.bosh /etc/network/interfaces && ifup -a'"))
		})
	})

	Describe("GetInterfaces", func() {
		Context("when a single dynamic network is used", func() {
			Context("and the virtual guest is only connected to the private network", func() {
				BeforeEach(func() {
					jsonBytes, err := json.Marshal(networkComponents)
					Expect(err).NotTo(HaveOccurred())

					softlayerClient.DoRawHttpRequestReturns(jsonBytes, 200, nil)
				})

				It("generates the correct interfaces", func() {
					interfaces, err := ubuntu.GetInterfaces(networks, 999)
					Expect(err).NotTo(HaveOccurred())

					Expect(interfaces).To(ConsistOf([]Interface{{
						Name:           "eth0",
						Auto:           true,
						AllowHotplug:   true,
						Address:        "10.155.248.190",
						Netmask:        "255.255.255.224",
						Gateway:        "10.155.248.161",
						DefaultGateway: true,
						Routes:         SoftlayerPrivateRoutes("10.155.248.161"),
					}}))

					Expect(softlayerClient.DoRawHttpRequestCallCount()).To(Equal(1))
					path, operation, body := softlayerClient.DoRawHttpRequestArgsForCall(0)
					Expect(path).To(Equal(`SoftLayer_Virtual_Guest/999/getObject?objectMask=mask[primaryBackendNetworkComponent.networkVlan.subnets,primaryNetworkComponent.networkVlan.subnets]`))
					Expect(operation).To(Equal("GET"))
					Expect(body.Len()).To(Equal(0))
				})
			})

			Context("and the virtual guest is connected to public and private networks", func() {
				BeforeEach(func() {
					networkComponents = VirtualGuestNetworkComponents{
						PrimaryBackendNetworkComponent: NetworkComponent{
							Name:             "eth",
							Port:             0,
							PrimaryIPAddress: "10.155.248.190",
							NetworkVLAN: NetworkVLAN{
								Name: "private vlan",
								Subnets: []Subnet{{
									NetworkIdentifier: "10.155.248.160",
									Gateway:           "10.155.248.161",
									BroadcastAddress:  "10.155.248.191",
									Netmask:           "255.255.255.224",
								}, {
									NetworkIdentifier: "10.155.198.0",
									Gateway:           "10.155.198.1",
									BroadcastAddress:  "10.155.198.63",
									Netmask:           "255.255.255.192",
								}, {
									NetworkIdentifier: "10.155.248.128",
									Gateway:           "10.155.248.129",
									BroadcastAddress:  "10.155.248.153",
									Netmask:           "255.255.255.240",
								}},
							},
						},
						PrimaryNetworkComponent: NetworkComponent{
							Name:             "eth",
							Port:             1,
							PrimaryIPAddress: "169.45.189.148",
							NetworkVLAN: NetworkVLAN{
								Name: "public vlan",
								Subnets: []Subnet{{
									NetworkIdentifier: "169.45.189.128",
									Gateway:           "169.45.189.129",
									BroadcastAddress:  "169.45.189.159",
									Netmask:           "255.255.255.224",
								}, {
									NetworkIdentifier: "169.45.188.208",
									Gateway:           "169.45.188.209",
									BroadcastAddress:  "169.45.188.223",
									Netmask:           "255.255.255.240",
								}},
							},
						},
					}

					jsonBytes, err := json.Marshal(networkComponents)
					Expect(err).NotTo(HaveOccurred())

					softlayerClient.DoRawHttpRequestReturns(jsonBytes, 200, nil)
				})

				It("generates the correct interfaces", func() {
					interfaces, err := ubuntu.GetInterfaces(networks, 999)
					Expect(err).NotTo(HaveOccurred())

					Expect(interfaces).To(ConsistOf([]Interface{{
						Name:         "eth0",
						Auto:         true,
						AllowHotplug: true,
						Address:      "10.155.248.190",
						Netmask:      "255.255.255.224",
						Gateway:      "10.155.248.161",
						Routes:       SoftlayerPrivateRoutes("10.155.248.161"),
					}, {
						Name:           "eth1",
						Auto:           true,
						AllowHotplug:   true,
						Address:        "169.45.189.148",
						Netmask:        "255.255.255.224",
						Gateway:        "169.45.189.129",
						DefaultGateway: true,
					}}))

					Expect(softlayerClient.DoRawHttpRequestCallCount()).To(Equal(1))
					path, operation, body := softlayerClient.DoRawHttpRequestArgsForCall(0)
					Expect(path).To(Equal(`SoftLayer_Virtual_Guest/999/getObject?objectMask=mask[primaryBackendNetworkComponent.networkVlan.subnets,primaryNetworkComponent.networkVlan.subnets]`))
					Expect(operation).To(Equal("GET"))
					Expect(body.Len()).To(Equal(0))
				})
			})

			Context("and manual networks are present", func() {
				BeforeEach(func() {
					networks = Networks{
						"dynamic": Network{
							Type:    "dynamic",
							Default: []string{"gateway"},
						},
						"private-manual": Network{
							Type:    "",
							IP:      "10.155.198.2",
							Netmask: "255.255.255.192",
							Gateway: "10.155.198.1",
						},
						"private-another-manual": Network{
							Type:    "manual",
							IP:      "10.155.248.130",
							Netmask: "255.255.255.240",
							Gateway: "10.155.248.129",
						},
						"public-manual": Network{
							Type:    "manual",
							IP:      "169.45.188.210",
							Netmask: "255.255.255.240",
							Gateway: "169.45.188.209",
						},
						"public-another-manual": Network{
							Type:    "",
							IP:      "169.45.188.211",
							Netmask: "255.255.255.240",
							Gateway: "169.45.188.209",
						},
					}

					networkComponents = VirtualGuestNetworkComponents{
						PrimaryBackendNetworkComponent: NetworkComponent{
							Name:             "eth",
							Port:             0,
							PrimaryIPAddress: "10.155.248.190",
							NetworkVLAN: NetworkVLAN{
								Name: "private vlan",
								Subnets: []Subnet{{
									NetworkIdentifier: "10.155.248.160",
									Gateway:           "10.155.248.161",
									BroadcastAddress:  "10.155.248.191",
									Netmask:           "255.255.255.224",
								}, {
									NetworkIdentifier: "10.155.198.0",
									Gateway:           "10.155.198.1",
									BroadcastAddress:  "10.155.198.63",
									Netmask:           "255.255.255.192",
								}, {
									NetworkIdentifier: "10.155.248.128",
									Gateway:           "10.155.248.129",
									BroadcastAddress:  "10.155.248.153",
									Netmask:           "255.255.255.240",
								}},
							},
						},
						PrimaryNetworkComponent: NetworkComponent{
							Name:             "eth",
							Port:             1,
							PrimaryIPAddress: "169.45.189.148",
							NetworkVLAN: NetworkVLAN{
								Name: "public vlan",
								Subnets: []Subnet{{
									NetworkIdentifier: "169.45.189.128",
									Gateway:           "169.45.189.129",
									BroadcastAddress:  "169.45.189.159",
									Netmask:           "255.255.255.224",
								}, {
									NetworkIdentifier: "169.45.188.208",
									Gateway:           "169.45.188.209",
									BroadcastAddress:  "169.45.188.223",
									Netmask:           "255.255.255.240",
								}},
							},
						},
					}

					jsonBytes, err := json.Marshal(networkComponents)
					Expect(err).NotTo(HaveOccurred())

					softlayerClient.DoRawHttpRequestReturns(jsonBytes, 200, nil)
				})

				It("generates all interfaces", func() {
					interfaces, err := ubuntu.GetInterfaces(networks, 999)
					Expect(err).NotTo(HaveOccurred())

					Expect(interfaces).To(ConsistOf([]Interface{{
						Name:           "eth0",
						Auto:           true,
						AllowHotplug:   true,
						Address:        "10.155.248.190",
						Netmask:        "255.255.255.224",
						Gateway:        "10.155.248.161",
						Routes:         SoftlayerPrivateRoutes("10.155.248.161"),
						DefaultGateway: false,
					}, {
						Name:           "eth1",
						Auto:           true,
						AllowHotplug:   true,
						Address:        "169.45.189.148",
						Netmask:        "255.255.255.224",
						Gateway:        "169.45.189.129",
						DefaultGateway: true,
					}, {
						Name:           "eth0:private-manual",
						Auto:           true,
						AllowHotplug:   true,
						Address:        "10.155.198.2",
						Netmask:        "255.255.255.192",
						Gateway:        "10.155.198.1",
						Routes:         SoftlayerPrivateRoutes("10.155.198.1"),
						DefaultGateway: false,
					}, {
						Name:           "eth0:private-another-manual",
						Auto:           true,
						AllowHotplug:   true,
						Address:        "10.155.248.130",
						Netmask:        "255.255.255.240",
						Gateway:        "10.155.248.129",
						Routes:         SoftlayerPrivateRoutes("10.155.248.129"),
						DefaultGateway: false,
					}, {
						Name:           "eth1:public-manual",
						Auto:           true,
						AllowHotplug:   true,
						Address:        "169.45.188.210",
						Netmask:        "255.255.255.240",
						Gateway:        "169.45.188.209",
						DefaultGateway: false,
					}, {
						Name:           "eth1:public-another-manual",
						Auto:           true,
						AllowHotplug:   true,
						Address:        "169.45.188.211",
						Netmask:        "255.255.255.240",
						Gateway:        "169.45.188.209",
						DefaultGateway: false,
					}}))

					Expect(softlayerClient.DoRawHttpRequestCallCount()).To(Equal(1))
					path, operation, body := softlayerClient.DoRawHttpRequestArgsForCall(0)
					Expect(path).To(Equal(`SoftLayer_Virtual_Guest/999/getObject?objectMask=mask[primaryBackendNetworkComponent.networkVlan.subnets,primaryNetworkComponent.networkVlan.subnets]`))
					Expect(operation).To(Equal("GET"))
					Expect(body.Len()).To(Equal(0))
				})
			})

			Context("and a private manual network is the default gateway", func() {
				BeforeEach(func() {
					networks = Networks{
						"dynamic": Network{
							Type: "dynamic",
						},
						"private-manual": Network{
							Type:    "",
							IP:      "10.155.198.2",
							Netmask: "255.255.255.192",
							Gateway: "10.155.198.1",
							Default: []string{"gateway"},
						},
					}

					networkComponents = VirtualGuestNetworkComponents{
						PrimaryBackendNetworkComponent: NetworkComponent{
							Name:             "eth",
							Port:             0,
							PrimaryIPAddress: "10.155.248.190",
							NetworkVLAN: NetworkVLAN{
								Name: "private vlan",
								Subnets: []Subnet{{
									NetworkIdentifier: "10.155.248.160",
									Gateway:           "10.155.248.161",
									BroadcastAddress:  "10.155.248.191",
									Netmask:           "255.255.255.224",
								}, {
									NetworkIdentifier: "10.155.198.0",
									Gateway:           "10.155.198.1",
									BroadcastAddress:  "10.155.198.63",
									Netmask:           "255.255.255.192",
								}},
							},
						},
					}

					jsonBytes, err := json.Marshal(networkComponents)
					Expect(err).NotTo(HaveOccurred())

					softlayerClient.DoRawHttpRequestReturns(jsonBytes, 200, nil)
				})

				It("generates the correct interfaces", func() {
					interfaces, err := ubuntu.GetInterfaces(networks, 999)
					Expect(err).NotTo(HaveOccurred())

					Expect(interfaces).To(ConsistOf([]Interface{{
						Name:           "eth0",
						Auto:           true,
						AllowHotplug:   true,
						Address:        "10.155.248.190",
						Netmask:        "255.255.255.224",
						Gateway:        "10.155.248.161",
						DefaultGateway: false,
						Routes:         SoftlayerPrivateRoutes("10.155.248.161"),
					}, {
						Name:           "eth0:private-manual",
						Auto:           true,
						AllowHotplug:   true,
						Address:        "10.155.198.2",
						Netmask:        "255.255.255.192",
						Gateway:        "10.155.198.1",
						DefaultGateway: true,
						Routes:         SoftlayerPrivateRoutes("10.155.198.1"),
					}}))

					Expect(softlayerClient.DoRawHttpRequestCallCount()).To(Equal(1))
					path, operation, body := softlayerClient.DoRawHttpRequestArgsForCall(0)
					Expect(path).To(Equal(`SoftLayer_Virtual_Guest/999/getObject?objectMask=mask[primaryBackendNetworkComponent.networkVlan.subnets,primaryNetworkComponent.networkVlan.subnets]`))
					Expect(operation).To(Equal("GET"))
					Expect(body.Len()).To(Equal(0))
				})
			})
		})

		Context("when multiple dynamic networks are used", func() {
			BeforeEach(func() {
				networks = Networks{
					"default": Network{
						Type:    "dynamic",
						Default: []string{"gateway"},
					},
					"dynamic": Network{
						Type: "dynamic",
					},
				}

				jsonBytes, err := json.Marshal(networkComponents)
				Expect(err).NotTo(HaveOccurred())

				softlayerClient.DoRawHttpRequestReturns(jsonBytes, 200, nil)
			})

			It("returns an error", func() {
				_, err := ubuntu.GetInterfaces(networks, 999)
				Expect(err).To(MatchError("virtual guests must have exactly one dynamic network"))
			})
		})

		Context("when an unknown network type is used", func() {
			BeforeEach(func() {
				networks = Networks{
					"default": Network{
						Type:    "dynamic",
						Default: []string{"gateway"},
					},
					"broken": Network{
						Type: "broken",
					},
				}

				jsonBytes, err := json.Marshal(networkComponents)
				Expect(err).NotTo(HaveOccurred())

				softlayerClient.DoRawHttpRequestReturns(jsonBytes, 200, nil)
			})

			It("returns an error", func() {
				_, err := ubuntu.GetInterfaces(networks, 999)
				Expect(err).To(MatchError("unexpected network type: broken"))
			})
		})
	})
})

var _ = Describe("Interfaces", func() {
	var interfaces Interfaces

	BeforeEach(func() {
		interfaces = Interfaces{}
	})

	Describe("Configuration", func() {
		It("generates a loopback configuration", func() {
			config, err := interfaces.Configuration()
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(BeEquivalentTo(LOOPBACK_ONLY))
		})

		Context("when an interface is not the default gateway", func() {
			BeforeEach(func() {
				interfaces = Interfaces{
					Interface{
						Name:    "interface-name",
						Address: "1.2.3.4",
						Netmask: "255.255.0.0",
					},
				}
			})

			It("generates a basic configuration", func() {
				config, err := interfaces.Configuration()
				Expect(err).NotTo(HaveOccurred())
				Expect(config).To(BeEquivalentTo(ONE_INTERFACE_NOT_DEFAULT_GATEWAY))
			})
		})

		Context("when an interface is present and the default gateway", func() {
			BeforeEach(func() {
				interfaces = Interfaces{
					Interface{
						Name:           "interface-name",
						Address:        "192.168.1.2",
						Netmask:        "255.255.255.0",
						Gateway:        "192.168.1.1",
						DefaultGateway: true,
					},
				}
			})

			It("generates a configuration with a gateway", func() {
				config, err := interfaces.Configuration()
				Expect(err).NotTo(HaveOccurred())
				Expect(config).To(BeEquivalentTo(ONE_INTERFACE_DEFAULT_GATEWAY))
			})
		})

		Context("when an interface contains contains routes", func() {
			BeforeEach(func() {
				interfaces = Interfaces{
					Interface{
						Name:    "interface-name",
						Address: "172.16.0.2",
						Netmask: "255.255.0.0",
						Routes: []Route{
							{Network: "192.168.1.0", Netmask: "255.255.255.0", Gateway: "172.16.0.1"},
							{Network: "10.0.0.0", Netmask: "255.0.0.0", Gateway: "172.16.0.1"},
						},
					},
				}
			})

			It("generates a configuration with routes", func() {
				config, err := interfaces.Configuration()
				Expect(err).NotTo(HaveOccurred())
				Expect(config).To(BeEquivalentTo(ONE_INTERFACE_WITH_ROUTES))
			})
		})

		Context("when multiple interfaces are present", func() {
			BeforeEach(func() {
				interfaces = Interfaces{
					Interface{
						Name:    "interface-name",
						Address: "1.2.3.4",
						Netmask: "255.255.0.0",
					},
					Interface{
						Name:    "another-interface-name",
						Address: "5.6.7.8",
						Netmask: "255.255.0.0",
					},
				}
			})

			It("generates a configuration for each interface", func() {
				config, err := interfaces.Configuration()
				Expect(err).NotTo(HaveOccurred())
				Expect(config).To(BeEquivalentTo(TWO_INTERFACES))
			})
		})
	})
})

const LOOPBACK_ONLY = `# Generated by softlayer-cpi
auto lo
iface lo inet loopback
`

const ONE_INTERFACE_NOT_DEFAULT_GATEWAY = `# Generated by softlayer-cpi
auto lo
iface lo inet loopback
# interface-name
iface interface-name inet static
    address 1.2.3.4
    netmask 255.255.0.0
`

const ONE_INTERFACE_DEFAULT_GATEWAY = `# Generated by softlayer-cpi
auto lo
iface lo inet loopback
# interface-name
iface interface-name inet static
    address 192.168.1.2
    netmask 255.255.255.0
    gateway 192.168.1.1
`

const ONE_INTERFACE_WITH_ROUTES = `# Generated by softlayer-cpi
auto lo
iface lo inet loopback
# interface-name
iface interface-name inet static
    address 172.16.0.2
    netmask 255.255.0.0
    post-up route add -net 192.168.1.0 netmask 255.255.255.0 gw 172.16.0.1
    post-up route add -net 10.0.0.0 netmask 255.0.0.0 gw 172.16.0.1
`

const TWO_INTERFACES = `# Generated by softlayer-cpi
auto lo
iface lo inet loopback
# interface-name
iface interface-name inet static
    address 1.2.3.4
    netmask 255.255.0.0
# another-interface-name
iface another-interface-name inet static
    address 5.6.7.8
    netmask 255.255.0.0
`
