package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	registryfakes "bosh-softlayer-cpi/registry/fakes"
	diskfakes "bosh-softlayer-cpi/softlayer/disk_service/fakes"
	imagefakes "bosh-softlayer-cpi/softlayer/stemcell_service/fakes"
	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"

	"bosh-softlayer-cpi/registry"
	boslconfig "bosh-softlayer-cpi/softlayer/config"
	"bosh-softlayer-cpi/softlayer/virtual_guest_service"

	"fmt"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
	"net"
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
			Mbus: "nats://nats:nats@fake-mbus:fake-port",
			Blobstore: registry.BlobstoreOptions{
				Provider: "dav",
				Options:  map[string]interface{}{"endpoint": "http://fake-blobstore:fake-port"},
			},
		}

		softlayerOptions = boslconfig.Config{
			Username:        "fake-username",
			ApiKey:          "fake-api-key",
			ApiEndpoint:     "fake-api-endpoint",
			DisableOsReload: false,
		}

		env = Environment(map[string]interface{}{"bosh": map[string]interface{}{"keep_root_password": false}})

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
				VmNamePrefix:      "fake-hostname",
				Domain:            "fake-domain.com",
				StartCpus:         2,
				MaxMemory:         2048,
				Datacenter:        "fake-datacenter",
				SshKey:            32345678,
				DeployedByBoshCLI: true,
			}

			networks = Networks{
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

			expectedInstanceNetworks = networks.AsInstanceServiceNetworks()

			expectedAgentSettings = registry.AgentSettings{
				AgentID: "fake-agent-id",
				Blobstore: registry.BlobstoreSettings{
					Provider: "dav",
					Options:  map[string]interface{}{"endpoint": "http://fake-blobstore:fake-port"},
				},
				Disks: registry.DisksSettings{
					System:     "",
					Persistent: map[string]registry.PersistentSettings{},
				},
				Mbus: "nats://nats:nats@fake-mbus:fake-port",
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
				Env: registry.EnvSettings(map[string]interface{}{
					"bosh": map[string]interface{}{
						"keep_root_password": true,
					},
				}),

				VM: registry.VMSettings{
					Name: "52345678",
				},
			}

			imageService.FindReturns(
				"12345678",
				nil,
			)
			vmService.GetVlanReturns(
				&datatypes.Network_Vlan{
					Id:           sl.Int(42345678),
					NetworkSpace: sl.String("PRIVATE"),
				},
				nil,
			)
			vmService.FindByPrimaryBackendIpReturns(
				&datatypes.Virtual_Guest{
					Id: sl.Int(52345678),
					Datacenter: &datatypes.Location{
						Name: sl.String("fake-datacenter-name"),
					},
				},
				nil,
			)
			vmService.ReloadOSReturns(
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

			vmService.FindReturns(
				&datatypes.Virtual_Guest{
					Id: sl.Int(52345678),
					Datacenter: &datatypes.Location{
						Name: sl.String("fake-datacenter-name"),
					},
					PrimaryBackendIpAddress:  sl.String("10.10.10.11"),
					FullyQualifiedDomainName: sl.String("fake-domain-name"),
				},
				nil,
			)
		})

		It("creates the vm when deployByBoshCli=true", func() {
			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).NotTo(HaveOccurred())
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(1))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeTrue())
			Expect(vmService.FindCallCount()).To(Equal(1))
			Expect(registryClient.UpdateSettings).To(Equal(expectedAgentSettings))
			actualCid, _ := vmService.ConfigureNetworksArgsForCall(0)
			Expect(vmCID).To(Equal(VMCID(actualCid).String()))
			_, actualInstanceNetworks := vmService.ConfigureNetworksArgsForCall(0)
			Expect(actualInstanceNetworks).To(Equal(expectedInstanceNetworks))
		})

		It("After creating the vm, /etc/hosts updated", func() {
			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).NotTo(HaveOccurred())
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(1))
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

		It("creates the vm when deployByBoshCli=false, and mbus host 0.0.0.0", func() {
			cloudProps = VMCloudProperties{
				VmNamePrefix: "fake-hostname",
				Domain:       "fake-domain.com",
				StartCpus:    2,
				MaxMemory:    2048,
				Datacenter:   "fake-datacenter",
				SshKey:       32345678,
			}

			agentOptions = registry.AgentOptions{
				Mbus: "nats://nats:nats@0.0.0.0:fake-port",
				Blobstore: registry.BlobstoreOptions{
					Provider: "dav",
					Options:  map[string]interface{}{"endpoint": "http://0.0.0.0:fake-port"},
				},
			}

			addrs, err := net.InterfaceAddrs()
			Expect(err).NotTo(HaveOccurred())

			var localIpAddr string

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						localIpAddr = ipnet.IP.String()
						break
					}
				}
			}

			expectedAgentSettings = registry.AgentSettings{
				AgentID: "fake-agent-id",
				Blobstore: registry.BlobstoreSettings{
					Provider: "dav",
					Options:  map[string]interface{}{"endpoint": fmt.Sprintf("http://%s:fake-port", localIpAddr)},
				},
				Disks: registry.DisksSettings{
					System:     "",
					Persistent: map[string]registry.PersistentSettings{},
				},
				Mbus: fmt.Sprintf("nats://nats:nats@%s:fake-port", localIpAddr),
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
				Env: registry.EnvSettings(map[string]interface{}{
					"bosh": map[string]interface{}{
						"keep_root_password": true,
					},
				}),

				VM: registry.VMSettings{
					Name: "52345678",
				},
			}

			createVM = NewCreateVM(
				imageService,
				vmService,
				registryClient,
				registryOptions,
				agentOptions,
				softlayerOptions,
			)

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).NotTo(HaveOccurred())
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(1))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeTrue())
			Expect(vmService.FindCallCount()).To(Equal(1))
			Expect(registryClient.UpdateSettings).To(Equal(expectedAgentSettings))
			actualCid, _ := vmService.ConfigureNetworksArgsForCall(0)
			Expect(vmCID).To(Equal(VMCID(actualCid).String()))
			_, actualInstanceNetworks := vmService.ConfigureNetworksArgsForCall(0)
			Expect(actualInstanceNetworks).To(Equal(expectedInstanceNetworks))
		})

		It("creates the vm when deployByBoshCli=false, and with mbus host not 0.0.0.0", func() {
			cloudProps = VMCloudProperties{
				VmNamePrefix: "fake-hostname",
				Domain:       "fake-domain.com",
				StartCpus:    2,
				MaxMemory:    2048,
				Datacenter:   "fake-datacenter",
				SshKey:       32345678,
			}

			agentOptions = registry.AgentOptions{
				Mbus: "nats://nats:nats@fake-mbus:fake-port",
				Blobstore: registry.BlobstoreOptions{
					Provider: "dav",
					Options:  map[string]interface{}{"endpoint": "http://fake-blobstore:fake-port"},
				},
			}

			networks = Networks{
				"fake-network-1": Network{
					Type:    "manual",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					CloudProperties: NetworkCloudProperties{
						VlanIds: []int{42345678},
					},
				},
				"fake-network-2": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanIds: []int{42345678},
					},
				},
			}

			expectedInstanceNetworks = networks.AsInstanceServiceNetworks()

			expectedAgentSettings = registry.AgentSettings{
				AgentID: "fake-agent-id",
				Blobstore: registry.BlobstoreSettings{
					Provider: "dav",
					Options:  map[string]interface{}{"endpoint": "http://fake-blobstore:fake-port"},
				},
				Disks: registry.DisksSettings{
					System:     "",
					Persistent: map[string]registry.PersistentSettings{},
				},
				Mbus: "nats://nats:nats@fake-mbus:fake-port",
				Networks: registry.NetworksSettings{
					"fake-network-1": registry.NetworkSettings{
						Type:    "manual",
						IP:      "10.10.10.10",
						Gateway: "fake-network-gateway",
						Netmask: "fake-network-netmask",
						DNS:     []string{"fake-network-dns"},
					},
					"fake-network-2": registry.NetworkSettings{
						Type:    "dynamic",
						IP:      "10.10.10.10",
						Gateway: "fake-network-gateway",
						Netmask: "fake-network-netmask",
						DNS:     []string{"fake-network-dns"},
						Default: []string{"fake-network-default"},
					},
				},
				Env: registry.EnvSettings(map[string]interface{}{
					"bosh": map[string]interface{}{
						"keep_root_password": true,
					},
				}),

				VM: registry.VMSettings{
					Name: "52345678",
				},
			}

			vmService.ConfigureNetworksReturns(
				instance.Networks{
					"fake-network-1": instance.Network{
						Type:    "manual",
						IP:      "10.10.10.10",
						Gateway: "fake-network-gateway",
						Netmask: "fake-network-netmask",
						DNS:     []string{"fake-network-dns"},
					},
					"fake-network-2": instance.Network{
						Type:    "dynamic",
						IP:      "10.10.10.10",
						Gateway: "fake-network-gateway",
						Netmask: "fake-network-netmask",
						DNS:     []string{"fake-network-dns"},
						Default: []string{"fake-network-default"},
					},
				},
				nil,
			)

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).NotTo(HaveOccurred())
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(1))
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeTrue())
			Expect(vmService.FindCallCount()).To(Equal(1))
			Expect(registryClient.UpdateSettings).To(Equal(expectedAgentSettings))
			actualCid, _ := vmService.ConfigureNetworksArgsForCall(0)
			Expect(vmCID).To(Equal(VMCID(actualCid).String()))
			_, actualInstanceNetworks := vmService.ConfigureNetworksArgsForCall(0)
			Expect(actualInstanceNetworks).To(Equal(expectedInstanceNetworks))
		})

		It("returns an error if imageService find call returns an error", func() {
			imageService.FindReturns(
				"12345678",
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
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if vmService get vlan call returns an error", func() {
			vmService.GetVlanReturns(
				&datatypes.Network_Vlan{},
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
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if vmService find by primary backend ip call returns an error", func() {
			vmService.FindByPrimaryBackendIpReturns(
				&datatypes.Virtual_Guest{},
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
			Expect(vmService.CreateCallCount()).To(Equal(0))
			Expect(vmService.ConfigureNetworksCallCount()).To(Equal(0))
			Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
			Expect(vmService.CleanUpCallCount()).To(Equal(0))
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if vmService reload os call returns an error", func() {
			vmService.ReloadOSReturns(
				errors.New("fake-vm-service-error"),
			)

			vmCID, err = createVM.Run(agentID, stemcellCID, cloudProps, networks, disks, env)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("OS reloading VM: fake-vm-service-error"))
			Expect(imageService.FindCallCount()).To(Equal(1))
			Expect(vmService.CreateSshKeyCallCount()).To(Equal(0))
			Expect(vmService.GetVlanCallCount()).To(Equal(1))
			Expect(vmService.FindByPrimaryBackendIpCallCount()).To(Equal(1))
			Expect(vmService.ReloadOSCallCount()).To(Equal(1))
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
						Provider: "dav",
						Options:  map[string]interface{}{"endpoint": "http://fake-blobstore:fake-port"},
					},
					Disks: registry.DisksSettings{
						System:     "",
						Persistent: map[string]registry.PersistentSettings{},
					},
					Mbus: "nats://nats:nats@fake-mbus:fake-port",
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
					Env: registry.EnvSettings(map[string]interface{}{
						"bosh": map[string]interface{}{
							"keep_root_password": true,
						},
					}),
					VM: registry.VMSettings{
						Name: "62345678",
					},
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
				Expect(vmService.CreateCallCount()).To(Equal(1))
				Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
				Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(0))
				Expect(vmService.CleanUpCallCount()).To(Equal(1))
				Expect(registryClient.UpdateCalled).To(BeTrue())
			})
		})

		Context("when cloud property options EphemeralDiskSize is set", func() {
			BeforeEach(func() {
				expectedAgentSettings = registry.AgentSettings{
					AgentID: "fake-agent-id",
					Blobstore: registry.BlobstoreSettings{
						Provider: "dav",
						Options:  map[string]interface{}{"endpoint": "http://fake-blobstore:fake-port"},
					},
					Disks: registry.DisksSettings{
						System:     "",
						Ephemeral:  "/dev/xvdc",
						Persistent: map[string]registry.PersistentSettings{},
					},
					Mbus: "nats://nats:nats@fake-mbus:fake-port",
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
					Env: registry.EnvSettings(map[string]interface{}{"bosh": map[string]interface{}{"keep_root_password": true}}),
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
					DeployedByBoshCLI: true,
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
				Expect(vmService.CreateCallCount()).To(Equal(0))
				Expect(vmService.ConfigureNetworksCallCount()).To(Equal(1))
				Expect(vmService.AttachEphemeralDiskCallCount()).To(Equal(1))
				Expect(registryClient.UpdateCalled).To(BeFalse())
				Expect(vmService.CleanUpCallCount()).To(Equal(0))
			})
		})
	})
})
