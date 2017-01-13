package vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"runtime"
	"time"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"

	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"
	fakesutil "github.com/cloudfoundry/bosh-softlayer-cpi/util/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	slh "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

var _ = Describe("SoftLayer_Virtual_Guest_Creator", func() {
	var (
		softLayerClient *fakeslclient.FakeSoftLayerClient
		sshClient       *fakesutil.FakeSshClient
		fakeVmFinder    *fakescommon.FakeVMFinder
		fakeVm          *fakescommon.FakeVM
		agentOptions    AgentOptions
		logger          boshlog.Logger
		creator         VMCreator
		featureOptions  FeatureOptions
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		sshClient = &fakesutil.FakeSshClient{}
		agentOptions = AgentOptions{Mbus: "fake-mbus"}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fakeVmFinder = &fakescommon.FakeVMFinder{}
		fakeVm = &fakescommon.FakeVM{}

		slh.TIMEOUT = 2 * time.Second
		slh.POLLING_INTERVAL = 1 * time.Second
	})

	Describe("#Create", func() {
		var (
			agentID    string
			stemcell   bslcstem.SoftLayerStemcell
			cloudProps VMCloudProperties
			networks   Networks
			env        Environment
		)

		Context("valid arguments with os_reload enabled", func() {
			BeforeEach(func() {
				agentID = "fake-agent-id"
				stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", softLayerClient, logger)

				env = Environment{}
				featureOptions = FeatureOptions{DisableOsReload: false}

				fakeVm.IDReturns(1234567)
				fakeVmFinder.FindReturns(fakeVm, true, nil)

				creator = NewSoftLayerCreator(
					fakeVmFinder,
					softLayerClient,
					agentOptions,
					logger,
					featureOptions,
				)
			})
			Context("creating vm by os_reload", func() {
				Context("with dynamic networking", func() {
					BeforeEach(func() {
						networks = map[string]Network{
							"fake-network0": Network{
								Type:    "dynamic",
								IP:      "10.0.0.11",
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
					})

					It("returns a new SoftLayerVM with ephemeral size", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							BoshIp:                       "10.0.0.1",
							EphemeralDiskSize:            25,
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-test",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
						}
						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize_OS_Reload(softLayerClient)

						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})

					It("returns a new SoftLayerVM without ephemeral size", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							BoshIp:                       "10.0.0.1",
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-test",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
						}
						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoflayerClientCreateObjectTestFixturesWithoutEphemeralDiskSize_OS_Reload(softLayerClient)
						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})
					It("returns a new SoftLayerVM without bosh ip", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							EphemeralDiskSize:            25,
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							NotDeployedByDirector: true,
						}

						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithoutBoshIP_OS_Reload(softLayerClient)

						switch runtime.GOOS {
						case "darwin":
							slh.NetworkInterface = "en0"
						case "linux":
							slh.NetworkInterface = "eth0"
						default:
							slh.NetworkInterface = "eth0"
						}
						slh.LocalDNSConfigurationFile = "/tmp/hosts"

						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})
					It("returns a new SoftLayerVM with neither bosh ip nor NotDeployedByDirector flag", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							EphemeralDiskSize:            25,
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
						}

						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithoutBoshIP_OS_Reload(softLayerClient)
						switch runtime.GOOS {
						case "darwin":
							slh.NetworkInterface = "en0"
						case "linux":
							slh.NetworkInterface = "eth0"
						default:
							slh.NetworkInterface = "eth0"
						}
						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})
				})
			})

			Context("creating vm in softlayer", func() {
				Context("without dynamic networking", func() {
					It("returns error", func() {
						networks = map[string]Network{
							"fake-network0": Network{
								Type:    "manual",
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

						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(vm).To(BeNil())
						Expect(err).To(Equal(bosherr.Error("virtual guests must have exactly one dynamic network")))
					})
				})

				Context("with dynamic networking", func() {
					BeforeEach(func() {
						networks = map[string]Network{
							"fake-network0": Network{
								Type:    "dynamic",
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
					})

					It("returns a new SoftLayerVM with ephemeral size", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							BoshIp:                       "10.0.0.1",
							EphemeralDiskSize:            25,
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-test",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
						}
						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize(softLayerClient)

						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})

					It("returns a new SoftLayerVM with ephemeral size", func() {
						networks = map[string]Network{
							"fake-network0": Network{
								Type:    "dynamic",
								IP:      "10.0.0.11",
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

						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							BoshIp:                       "10.0.0.1",
							EphemeralDiskSize:            25,
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-test",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							DisableOsReload: true,
						}
						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize(softLayerClient)

						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})

					It("returns a new SoftLayerVM without ephemeral size", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							BoshIp:                       "10.0.0.1",
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-test",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
						}
						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithoutEphemeralDiskSize(softLayerClient)
						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})
					It("returns a new SoftLayerVM without bosh ip", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							EphemeralDiskSize:            25,
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							NotDeployedByDirector: true,
						}

						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithoutBoshIP(softLayerClient)

						switch runtime.GOOS {
						case "darwin":
							slh.NetworkInterface = "en0"
						case "linux":
							slh.NetworkInterface = "eth0"
						default:
							slh.NetworkInterface = "eth0"
						}
						slh.LocalDNSConfigurationFile = "/tmp/hosts"

						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})
					It("returns a new SoftLayerVM with neither bosh ip nor NotDeployedByDirector flag", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							EphemeralDiskSize:            25,
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
						}

						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithoutBoshIP(softLayerClient)
						switch runtime.GOOS {
						case "darwin":
							slh.NetworkInterface = "en0"
						case "linux":
							slh.NetworkInterface = "eth0"
						default:
							slh.NetworkInterface = "eth0"
						}
						slh.LocalDNSConfigurationFile = "/tmp/hosts"

						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})
				})
			})
		})

		Context("valid arguments with os_reload disabled", func() {
			BeforeEach(func() {
				agentID = "fake-agent-id"
				stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", softLayerClient, logger)

				env = Environment{}

				featureOptions = FeatureOptions{DisableOsReload: true}

				fakeVm.IDReturns(1234567)
				fakeVmFinder.FindReturns(fakeVm, true, nil)

				creator = NewSoftLayerCreator(
					fakeVmFinder,
					softLayerClient,
					agentOptions,
					logger,
					featureOptions,
				)
			})

			Context("creating vm in softlayer", func() {
				Context("with dynamic networking", func() {
					BeforeEach(func() {
						networks = map[string]Network{
							"fake-network0": Network{
								Type:    "dynamic",
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
					})

					It("returns a new SoftLayerVM with ephemeral size", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							BoshIp:                       "10.0.0.1",
							EphemeralDiskSize:            25,
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-test",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
						}
						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize(softLayerClient)

						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})
					It("returns a new SoftLayerVM without ephemeral size", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							BoshIp:                       "10.0.0.1",
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-test",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
						}
						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithoutEphemeralDiskSize(softLayerClient)
						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})
					It("returns a new SoftLayerVM without bosh ip", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							EphemeralDiskSize:            25,
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							DisableOsReload:       true,
							NotDeployedByDirector: true,
						}

						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithoutBoshIP(softLayerClient)

						switch runtime.GOOS {
						case "darwin":
							slh.NetworkInterface = "en0"
						case "linux":
							slh.NetworkInterface = "eth0"
						default:
							slh.NetworkInterface = "eth0"
						}
						slh.LocalDNSConfigurationFile = "/tmp/hosts"

						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})
					It("returns a new SoftLayerVM with neither bosh ip nor NotDeployedByDirector flag", func() {
						cloudProps = VMCloudProperties{
							StartCpus: 4,
							MaxMemory: 2048,
							Domain:    "fake-domain.com",
							BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
								GlobalIdentifier: "fake-uuid",
							},
							RootDiskSize:                 25,
							EphemeralDiskSize:            25,
							Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
							HourlyBillingFlag:            true,
							LocalDiskFlag:                true,
							VmNamePrefix:                 "bosh-",
							PostInstallScriptUri:         "",
							DedicatedAccountHostOnlyFlag: true,
							PrivateNetworkOnlyFlag:       false,
							SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
							BlockDevices: []sldatatypes.BlockDevice{{
								Device:    "0",
								DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
							NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
							PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
								NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
							DisableOsReload: true,
						}

						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}
						setFakeSoftlayerClientCreateObjectTestFixturesWithoutBoshIP(softLayerClient)

						switch runtime.GOOS {
						case "darwin":
							slh.NetworkInterface = "en0"
						case "linux":
							slh.NetworkInterface = "eth0"
						default:
							slh.NetworkInterface = "eth0"
						}

						slh.LocalDNSConfigurationFile = "/tmp/hosts"
						_, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})
})

func setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_createObject.json",

		"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getUpgradeItemPrices.json",
		"SoftLayer_Virtual_Guest_Service_getLocalDiskFlag_local.json",
		"SoftLayer_Product_Order_Service_placeOrder.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getLastTransaction_CloudInstanceUpgrade.json",
		"SoftLayer_Virtual_Guest_Service_getPowerState.json",
		"SoftLayer_Virtual_Guest_Service_getBlockDevices.json",

		"SoftLayer_Virtual_Guest_Service_getObject.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}

func setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize_OS_Reload(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_getObjects.json",
		"SoftLayer_Virtual_Guest_Service_editObject.json",
		"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getUpgradeItemPrices.json",
		"SoftLayer_Virtual_Guest_Service_getLocalDiskFlag_local.json",
		"SoftLayer_Product_Order_Service_placeOrder.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getLastTransaction_CloudInstanceUpgrade.json",
		"SoftLayer_Virtual_Guest_Service_getPowerState.json",
		"SoftLayer_Virtual_Guest_Service_getBlockDevices.json",

		"SoftLayer_Virtual_Guest_Service_getObject.json",
		"SoftLayer_Virtual_Guest_Service_getObject.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}

func setFakeSoftlayerClientCreateObjectTestFixturesWithoutEphemeralDiskSize(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_createObject.json",

		"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",

		"SoftLayer_Virtual_Guest_Service_getObject.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}

func setFakeSoflayerClientCreateObjectTestFixturesWithoutEphemeralDiskSize_OS_Reload(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_getObjects.json",
		"SoftLayer_Virtual_Guest_Service_editObject.json",

		"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",

		"SoftLayer_Virtual_Guest_Service_getObject.json",
		"SoftLayer_Virtual_Guest_Service_getObject.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}

func setFakeSoftlayerClientCreateObjectTestFixturesWithoutBoshIP(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_createObject.json",
		"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getUpgradeItemPrices.json",
		"SoftLayer_Virtual_Guest_Service_getLocalDiskFlag_local.json",
		"SoftLayer_Product_Order_Service_placeOrder.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getLastTransaction_CloudInstanceUpgrade.json",
		"SoftLayer_Virtual_Guest_Service_getPowerState.json",
		"SoftLayer_Virtual_Guest_Service_getBlockDevices.json",

		"SoftLayer_Virtual_Guest_Service_getObject.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}

func setFakeSoftlayerClientCreateObjectTestFixturesWithoutBoshIP_OS_Reload(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_getObjects.json",
		"SoftLayer_Virtual_Guest_Service_editObject.json",

		"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getUpgradeItemPrices.json",
		"SoftLayer_Virtual_Guest_Service_getLocalDiskFlag_local.json",
		"SoftLayer_Product_Order_Service_placeOrder.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getLastTransaction_CloudInstanceUpgrade.json",
		"SoftLayer_Virtual_Guest_Service_getPowerState.json",
		"SoftLayer_Virtual_Guest_Service_getBlockDevices.json",

		"SoftLayer_Virtual_Guest_Service_getObject.json",
		"SoftLayer_Virtual_Guest_Service_getObject.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}
