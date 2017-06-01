package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	diskfakes "bosh-softlayer-cpi/softlayer/disk_service/fakes"
	imagefakes "bosh-softlayer-cpi/softlayer/stemcell_service/fakes"
	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"
	registryfakes "bosh-softlayer-cpi/registry/fakes"

	"bosh-softlayer-cpi/softlayer/virtual_guest_service"
	"bosh-softlayer-cpi/registry"
	boslconfig "bosh-softlayer-cpi/softlayer/config"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("CreateVM", func() {
	var (
		err                      error
		vmCID                    string
		agentID                  string
		stemcellCID              StemcellCID
		disks                    []DiskCID
		env                      Environment
		networks                 Networks
		cloudProps               VMCloudProperties
		registryOptions          registry.ClientOptions
		agentOptions             registry.AgentOptions
		softlayerOptions         boslconfig.Config
		expectedVMProps          *instance.Properties
		expectedInstanceNetworks instance.Networks
		expectedAgentSettings    registry.AgentSettings

		vmService      *instancefakes.FakeService
		diskService    *diskfakes.FakeService
		imageService   *imagefakes.FakeService
		registryClient *registryfakes.FakeClient

		createVM CreateVM
	)

	BeforeEach(func() {
		vmService = &instancefakes.FakeService{}
		diskService = &diskfakes.FakeService{}
		imageService = &imagefakes.FakeService{}
		registryClient = &registryfakes.FakeClient{}
		registryOptions = registry.ClientOptions{
			Protocol: "http",
			Host:     "fake-registry-host",
			Port:     25777,
			Username: "fake-registry-username",
			Password: "fake-registry-password",
			HTTPOptions: registry.HttpRegistryOptions{
				Port:     25777,
				User:     "fake-registry-username",
				Password: "fake-registry-password",
			},
		}
		agentOptions = registry.AgentOptions{
			Mbus: "http://fake-mbus",
			Blobstore: registry.BlobstoreOptions{
				Provider: "fake-blobstore-type",
			},
		}

		softlayerOptions = boslconfig.Config{
			Username:        "fake-username",
			ApiKey:          "fake-api-key",
			ApiEndpoint:     "fake-api-endpoint",
			DisableOsReload: false,
		}

		createVM = NewCreateVM(
			imageService,
			vmService,
			registryClient,
			registryOptions,
			agentOptions,
			softlayerOptions,
		)
	})

	Describe("Run", func() {
		BeforeEach(func() {
			agentID = "fake-agent-id"
			stemcellCID = StemcellCID(12345678)

			cloudProps = VMCloudProperties{
				VmNamePrefix: "fake-hostname",
				Domain:       "fake-domain.com",
				StartCpus:    2,
				MaxMemory:    2048,
				Datacenter:   "fake-datacenter",
				SshKey:       32345678,
			}

			networks = Networks{
				"fake-network-name": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					DHCP:    true,
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              42345678,
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}

			expectedVMProps = &instance.Properties{
				VirtualGuestTemplate: datatypes.Virtual_Guest{
					Id: sl.Int(52345678),
					Datacenter: &datatypes.Location{
						Name: sl.String("fake-datacenter-name"),
					},
				},
				DeployedByBoshCLI: true,
			}

			expectedInstanceNetworks = networks.AsInstanceServiceNetworks()
			expectedAgentSettings = registry.AgentSettings{
				AgentID: "fake-agent-id",
				Blobstore: registry.BlobstoreSettings{
					Provider: "fake-blobstore-type",
				},
				Disks: registry.DisksSettings{
					System:     "",
					Persistent: map[string]registry.PersistentSettings{},
				},
				Mbus: "http://fake-mbus",
				Networks: registry.NetworksSettings{
					"fake-network-name": registry.NetworkSettings{
						Type:    "dynamic",
						IP:      "10.10.10.10",
						Gateway: "fake-network-gateway",
						Netmask: "fake-network-netmask",
						DNS:     []string{"fake-network-dns"},
						Default: []string{"fake-network-default"},
					},
				},
				VM: registry.VMSettings{
					Name: "52345678",
				},
			}

			imageService.FindReturns(
				"12345678",
				true,
				nil,
			)
			vmService.GetVlanReturns(
				datatypes.Network_Vlan{
					Id:           sl.Int(42345678),
					NetworkSpace: sl.String("PRIVATE"),
				},
				nil,
			)
			vmService.FindByPrimaryBackendIpReturns(
				datatypes.Virtual_Guest{
					Id: sl.Int(52345678),
					Datacenter: &datatypes.Location{
						Name: sl.String("fake-datacenter-name"),
					},
				},
				true,
				nil,
			)
			vmService.ReloadOSReturns(
				"fake-token",
				nil,
			)
			vmService.EditReturns(
				true,
				nil,
			)
			vmService.ConfigureNetworksReturns(
				instance.Networks{
					"fake-network-name": instance.Network{
						Type:    "dynamic",
						IP:      "10.10.10.10",
						Gateway: "fake-network-gateway",
						Netmask: "fake-network-netmask",
						DNS:     []string{"fake-network-dns"},
						Default: []string{"fake-network-default"},
						CloudProperties: instance.NetworkCloudProperties{
							VlanID:              42345678,
							SourcePolicyRouting: true,
							Tags:                []string{"fake-network-cloud-network-tag"},
						},
					},
				},
				nil,
			)

		})

		// AttachEphemeralDiskCallCount

		It("creates the vm", func() {
			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).NotTo(HaveOccurred())
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(1))
			Expect(vmService.EditCallCount()).To(Equal(1))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeTrue())
			Expect(registryClient.UpdateSettings).To(Equal(expectedAgentSettings))
			actualCid, _ := vmService.ConfigureNetworksArgsForCall(0)
			Expect(vmCID).To(Equal(VMCID(actualCid).String()))
			_, actualInstanceNetworks := vmService.ConfigureNetworksArgsForCall(0)
			Expect(actualInstanceNetworks).To(Equal(expectedInstanceNetworks))
		})

		It("returns an error if imageService find call returns an error", func() {
			imageService.FindReturns(
				"12345678",
				false,
				errors.New("fake-image-service-error"),
			)

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-image-service-error"))
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(0))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(0))
			Expect(vmService.ReloadOSCallCount()).To(Equal(0))
			Expect(vmService.EditCallCount()).To(Equal(0))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if stemcell is not found", func() {
			imageService.FindReturns(
				"12345678",
				false,
				nil,
			)

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Stemcell '12345678' not found"))
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(0))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(0))
			Expect(vmService.ReloadOSCallCount()).To(Equal(0))
			Expect(vmService.EditCallCount()).To(Equal(0))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if vmService get vlan call returns an error", func() {
			vmService.GetVlanReturns(
				datatypes.Network_Vlan{},
				errors.New("fake-vm-service-error"),
			)

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(0))
			Expect(vmService.ReloadOSCallCount()).To(Equal(0))
			Expect(vmService.EditCallCount()).To(Equal(0))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if vmService find by primary backend ip call returns an error", func() {
			vmService.FindByPrimaryBackendIpReturns(
				datatypes.Virtual_Guest{},
				false,
				errors.New("fake-vm-service-error"),
			)

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(0))
			Expect(vmService.EditCallCount()).To(Equal(0))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if vm with IP address is not found", func() {
			vmService.FindByPrimaryBackendIpReturns(
				datatypes.Virtual_Guest{},
				false,
				nil,
			)

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Finding VM with IP Address '10.10.10.10'"))
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(0))
			Expect(vmService.EditCallCount()).To(Equal(0))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if vmService reload os call returns an error", func() {
			vmService.ReloadOSReturns(
				"",
				errors.New("fake-vm-service-error"),
			)

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to do OS Reload with IP Address '10.10.10.10' with error"))
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(1))
			Expect(vmService.EditCallCount()).To(Equal(0))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if vmService edit returns an error", func() {
			vmService.EditReturns(
				false,
				errors.New("fake-vm-service-error"),
			)

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(1))
			Expect(vmService.EditCallCount()).To(Equal(1))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if vmService configure networks returns an error", func() {
			vmService.ConfigureNetworksReturns(
				instance.Networks{},
				errors.New("fake-vm-service-error"),
			)

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(1))
			Expect(vmService.EditCallCount()).To(Equal(1))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if registryClient update call returns an error", func() {
			registryClient.UpdateErr = errors.New("fake-registry-client-error")

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-registry-client-error"))
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(1))
			Expect(vmService.EditCallCount()).To(Equal(1))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeTrue())
		})

		Context("when softlayer options PublicKey is set", func() {
			BeforeEach(func() {
				softlayerOptions = boslconfig.Config{
					Username:             "fake-username",
					ApiKey:               "fake-api-key",
					ApiEndpoint:          "fake-api-endpoint",
					DisableOsReload:      false,
					PublicKey:            "fake-public-key",
					PublicKeyFingerPrint: "fake-public-key-fingerprint",
				}

				createVM = NewCreateVM(
					imageService,
					vmService,
					registryClient,
					registryOptions,
					agentOptions,
					softlayerOptions,
				)
			})

			It("creates the vm and create ssh key", func() {
				vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
				Expect(err).NotTo(HaveOccurred())
				Expect(imageService.FindCallCount()).To(Equal(1))
				Expect(vmService.CreateSshKeyCallCount()).To(Equal(1))
				Expect(vmService.GetVlanCallCount()).To(Equal(1))
				Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
				Expect(vmService.ReloadOSCallCount()).To(Equal(1))
				Expect(vmService.EditCallCount()).To(Equal(1))
				Expect(vmService.CreateCallCount()).To(Equal(0))
				Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
				Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
				Expect(vmService.CleanUpCallCount()).To(Equal(0))
				Expect(registryClient.UpdateCalled).To(BeTrue())
				Expect(registryClient.UpdateSettings).To(Equal(expectedAgentSettings))
				actualCid, _ := vmService.ConfigureNetworksArgsForCall(0)
				Expect(vmCID).To(Equal(VMCID(actualCid).String()))
				_, actualInstanceNetworks := vmService.ConfigureNetworksArgsForCall(0)
				Expect(actualInstanceNetworks).To(Equal(expectedInstanceNetworks))
			})

			It("returns an error if vmService create ssh key call returns an error", func() {
				vmService.CreateSshKeyReturns(
					0,
					errors.New("fake-vm-service-error"),
				)

				vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
				Expect(imageService.FindCallCount()).To(Equal(1))
				Expect(vmService.CreateSshKeyCallCount()).To(Equal(1))
				Expect(vmService.GetVlanCallCount()).To(Equal(0))
				Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(0))
				Expect(vmService.ReloadOSCallCount()).To(Equal(0))
				Expect(vmService.EditCallCount()).To(Equal(0))
				Expect(vmService.CreateCallCount()).To(Equal(0))
				Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
				Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
				Expect(vmService.CleanUpCallCount()).To(Equal(0))
				Expect(registryClient.UpdateCalled).To(BeFalse())
			})
		})

		Context("when softlayer options DisableOsReload is set", func() {
			BeforeEach(func() {
				softlayerOptions = boslconfig.Config{
					Username:        "fake-username",
					ApiKey:          "fake-api-key",
					ApiEndpoint:     "fake-api-endpoint",
					DisableOsReload: true,
				}

				createVM = NewCreateVM(
					imageService,
					vmService,
					registryClient,
					registryOptions,
					agentOptions,
					softlayerOptions,
				)

				expectedAgentSettings = registry.AgentSettings{
					AgentID: "fake-agent-id",
					Blobstore: registry.BlobstoreSettings{
						Provider: "fake-blobstore-type",
					},
					Disks: registry.DisksSettings{
						System:     "",
						Persistent: map[string]registry.PersistentSettings{},
					},
					Mbus: "http://fake-mbus",
					Networks: registry.NetworksSettings{
						"fake-network-name": registry.NetworkSettings{
							Type:    "dynamic",
							IP:      "10.10.10.10",
							Gateway: "fake-network-gateway",
							Netmask: "fake-network-netmask",
							DNS:     []string{"fake-network-dns"},
							Default: []string{"fake-network-default"},
						},
					},
					VM: registry.VMSettings{
						Name: "62345678",
					},
				}
				expectedVMProps = &instance.Properties{
					VirtualGuestTemplate: datatypes.Virtual_Guest{
						Id: sl.Int(52345678),
						Datacenter: &datatypes.Location{
							Name: sl.String("fake-datacenter-name"),
						},
						BlockDeviceTemplateGroup: &datatypes.Virtual_Guest_Block_Device_Template_Group{},
						Hostname:                 sl.String("fake-hostname"),
					},
					DeployedByBoshCLI: true,
				}

				vmService.CreateReturns(
					62345678,
					nil,
				)
			})

			It("creates the vm", func() {
				vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
				Expect(err).NotTo(HaveOccurred())
				Expect(imageService.FindCallCount()).To(Equal(1))
				Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
				Expect(vmService.GetVlanCallCount()).To(Equal(1))
				Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(0))
				Expect(vmService.ReloadOSCallCount()).To(Equal(0))
				Expect(vmService.EditCallCount()).To(Equal(0))
				Expect(vmService.CreateCallCount()).To(Equal(1))
				Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
				Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
				Expect(vmService.CleanUpCallCount()).To(Equal(0))
				Expect(registryClient.UpdateCalled).To(BeTrue())
				Expect(registryClient.UpdateSettings).To(Equal(expectedAgentSettings))
				actualCid, _ := vmService.ConfigureNetworksArgsForCall(0)
				Expect(vmCID).To(Equal(VMCID(actualCid).String()))
				_, actualInstanceNetworks := vmService.ConfigureNetworksArgsForCall(0)
				Expect(actualInstanceNetworks).To(Equal(expectedInstanceNetworks))
				_, _, actualRegistryEndpoint := vmService.CreateArgsForCall(0)
				Expect(actualRegistryEndpoint).To(Equal("http://fake-registry-username:fake-registry-password@fake-registry-host:25777"))
			})

			It("returns an error if vmService create call returns an error", func() {
				vmService.CreateReturns(
					0,
					errors.New("fake-vm-service-error"),
				)

				vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
				Expect(imageService.FindCallCount()).To(Equal(1))
				Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
				Expect(vmService.GetVlanCallCount()).To(Equal(1))
				Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(0))
				Expect(vmService.ReloadOSCallCount()).To(Equal(0))
				Expect(vmService.EditCallCount()).To(Equal(0))
				Expect(vmService.CreateCallCount()).To(Equal(1))
				Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
				Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
				Expect(vmService.CleanUpCallCount()).To(Equal(0))
				Expect(registryClient.UpdateCalled).To(BeFalse())
			})

			It("returns an error and cleans up if vmService configure networks returns an error", func() {
				vmService.ConfigureNetworksReturns(
					instance.Networks{},
					errors.New("fake-vm-service-error"),
				)

				vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
				Expect(imageService.FindCallCount()).To(Equal(1))
				Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
				Expect(vmService.GetVlanCallCount()).To(Equal(1))
				Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(0))
				Expect(vmService.ReloadOSCallCount()).To(Equal(0))
				Expect(vmService.EditCallCount()).To(Equal(0))
				Expect(vmService.CreateCallCount()).To(Equal(1))
				Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
				Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
				Expect(vmService.CleanUpCallCount()).To(Equal(1))
				Expect(registryClient.UpdateCalled).To(BeFalse())
			})

			It("returns an error and cleans up if registryClient update call returns an error", func() {
				registryClient.UpdateErr = errors.New("fake-registry-client-error")

				vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-registry-client-error"))
				Expect(imageService.FindCallCount()).To(Equal(1))
				Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
				Expect(vmService.GetVlanCallCount()).To(Equal(1))
				Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(0))
				Expect(vmService.ReloadOSCallCount()).To(Equal(0))
				Expect(vmService.EditCallCount()).To(Equal(0))
				Expect(vmService.CreateCallCount()).To(Equal(1))
				Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
				Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
				Expect(vmService.CleanUpCallCount()).To(Equal(1))
				Expect(registryClient.UpdateCalled).To(BeTrue())
			})
		})

		Context("when cloud propertys options EphemeralDiskSize is set", func() {
			BeforeEach(func() {
				expectedAgentSettings = registry.AgentSettings{
					AgentID: "fake-agent-id",
					Blobstore: registry.BlobstoreSettings{
						Provider: "fake-blobstore-type",
					},
					Disks: registry.DisksSettings{
						System:     "",
						Ephemeral:  "/dev/xvdc",
						Persistent: map[string]registry.PersistentSettings{},
					},
					Mbus: "http://fake-mbus",
					Networks: registry.NetworksSettings{
						"fake-network-name": registry.NetworkSettings{
							Type:    "dynamic",
							IP:      "10.10.10.10",
							Gateway: "fake-network-gateway",
							Netmask: "fake-network-netmask",
							DNS:     []string{"fake-network-dns"},
							Default: []string{"fake-network-default"},
						},
					},
					VM: registry.VMSettings{
						Name: "52345678",
					},
				}

				cloudProps = VMCloudProperties{
					VmNamePrefix:      "fake-hostname",
					Domain:            "fake-domain.com",
					StartCpus:         2,
					MaxMemory:         2048,
					Datacenter:        "fake-datacenter",
					SshKey:            32345678,
					EphemeralDiskSize: 2048,
				}
			})

			It("creates the vm and attaches ephemeral disk", func() {
				vmService.AttachEphemeralDiskReturns(
					nil,
				)

				vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
				Expect(err).NotTo(HaveOccurred())
				Expect(imageService.FindCallCount()).To(Equal(1))
				Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
				Expect(vmService.GetVlanCallCount()).To(Equal(1))
				Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
				Expect(vmService.ReloadOSCallCount()).To(Equal(1))
				Expect(vmService.EditCallCount()).To(Equal(1))
				Expect(vmService.CreateCallCount()).To(Equal(0))
				Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
				Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(1))
				Expect(registryClient.UpdateCalled).To(BeTrue())
				Expect(vmService.CleanUpCallCount()).To(Equal(0))
				Expect(registryClient.UpdateSettings).To(Equal(expectedAgentSettings))
				actualCid, _ := vmService.ConfigureNetworksArgsForCall(0)
				Expect(vmCID).To(Equal(VMCID(actualCid).String()))
				_, actualInstanceNetworks := vmService.ConfigureNetworksArgsForCall(0)
				Expect(actualInstanceNetworks).To(Equal(expectedInstanceNetworks))
			})

			It("returns an error if vmService attach ephemeral disk call returns an error", func() {
				vmService.AttachEphemeralDiskReturns(
					errors.New("fake-vm-service-error"),
				)

				vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
				Expect(imageService.FindCallCount()).To(Equal(1))
				Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
				Expect(vmService.GetVlanCallCount()).To(Equal(1))
				Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
				Expect(vmService.ReloadOSCallCount()).To(Equal(1))
				Expect(vmService.EditCallCount()).To(Equal(1))
				Expect(vmService.CreateCallCount()).To(Equal(0))
				Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
				Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(1))
				Expect(registryClient.UpdateCalled).To(BeFalse())
				Expect(vmService.CleanUpCallCount()).To(Equal(0))
			})
		})
	})
})
