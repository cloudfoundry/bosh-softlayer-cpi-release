package vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"time"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	testhelpers "github.com/maximilien/bosh-softlayer-cpi/test_helpers"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	fakestem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell/fakes"
	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
	fakesutil "github.com/maximilien/bosh-softlayer-cpi/util/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	bslcommon "github.com/maximilien/bosh-softlayer-cpi/softlayer/common"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

var _ = Describe("SoftLayerCreator", func() {
	var (
		softLayerClient        *fakeslclient.FakeSoftLayerClient
		sshClient              *fakesutil.FakeSshClient
		agentEnvServiceFactory *fakevm.FakeAgentEnvServiceFactory
		fs                     *fakesys.FakeFileSystem
		uuidGenerator          *fakeuuid.FakeGenerator
		agentOptions           AgentOptions
		logger                 boshlog.Logger
		creator                SoftLayerCreator
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		sshClient = fakesutil.NewFakeSshClient()
		uuidGenerator = fakeuuid.NewFakeGenerator()
		fs = fakesys.NewFakeFileSystem()
		agentEnvServiceFactory = &fakevm.FakeAgentEnvServiceFactory{}
		agentOptions = AgentOptions{Mbus: "fake-mbus"}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		creator = NewSoftLayerCreator(
			softLayerClient,
			agentEnvServiceFactory,
			agentOptions,
			logger,
			uuidGenerator,
			fs,
		)
		bslcommon.TIMEOUT = 2 * time.Second
		bslcommon.POLLING_INTERVAL = 1 * time.Second

		os.Setenv("OS_RELOAD_ENABLED", "FALSE")
		os.Setenv("SQLITE_DB_FOLDER", "/tmp")
	})

	Describe("#Create", func() {
		var (
			agentID    string
			stemcell   bslcstem.SoftLayerStemcell
			cloudProps VMCloudProperties
			networks   Networks
			env        Environment
		)

		Context("valid arguments", func() {
			BeforeEach(func() {
				agentID = "fake-agent-id"
				stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", fakestem.FakeStemcellKind, softLayerClient, logger)
				networks = Networks{}
				env = Environment{}

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
				testhelpers.SetTestFixturesForFakeSSHClient(sshClient, expectedCmdResults, nil)
				setFakeSoftLayerClientCreateObjectTestFixturesWithEphemeralDiskSize(softLayerClient)
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
				testhelpers.SetTestFixturesForFakeSSHClient(sshClient, expectedCmdResults, nil)
				setFakeSoftLayerClientCreateObjectTestFixturesWithoutEphemeralDiskSize(softLayerClient)
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
				}

				expectedCmdResults := []string{
					"",
				}
				testhelpers.SetTestFixturesForFakeSSHClient(sshClient, expectedCmdResults, nil)
				setFakeSoftLayerClientCreateObjectTestFixturesWithoutBoshIP(softLayerClient)
				vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
				Expect(err).ToNot(HaveOccurred())
				Expect(vm.ID()).To(Equal(1234567))
			})
		})

		Context("invalid arguments", func() {
			Context("missing correct VMProperties", func() {
				BeforeEach(func() {
					agentID = "fake-agent-id"
					stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", fakestem.FakeStemcellKind, softLayerClient, logger)
					networks = Networks{}
					env = Environment{}

					setFakeSoftLayerClientCreateObjectTestFixturesWithEphemeralDiskSize(softLayerClient)
				})

				It("fails when VMProperties is missing StartCpus", func() {
					cloudProps = VMCloudProperties{
						MaxMemory:  2048,
						Datacenter: sldatatypes.Datacenter{Name: "fake-datacenter"},
					}

					_, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
					Expect(err).To(HaveOccurred())
				})

				It("fails when VMProperties is missing MaxMemory", func() {
					cloudProps = VMCloudProperties{
						StartCpus:  4,
						Datacenter: sldatatypes.Datacenter{Name: "fake-datacenter"},
					}

					_, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
					Expect(err).To(HaveOccurred())
				})

				It("fails when VMProperties is missing Domain", func() {
					cloudProps = VMCloudProperties{
						StartCpus: 4,
						MaxMemory: 1024,
					}

					_, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})

func setFakeSoftLayerClientCreateObjectTestFixturesWithEphemeralDiskSize(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_createObject.json",
		"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getUpgradeItemPrices.json",
		"SoftLayer_Product_Order_Service_placeOrder.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getLastTransaction_CloudInstanceUpgrade.json",
		"SoftLayer_Virtual_Guest_Service_getPowerState.json",

		"SoftLayer_Virtual_Guest_Service_getObject.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}

func setFakeSoftLayerClientCreateObjectTestFixturesWithoutEphemeralDiskSize(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_createObject.json",

		"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",

		"SoftLayer_Virtual_Guest_Service_getObject.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}

func setFakeSoftLayerClientCreateObjectTestFixturesWithoutBoshIP(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_createObject.json",
		"SoftLayer_Virtual_Guest_Service_getLastTransaction.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getUpgradeItemPrices.json",
		"SoftLayer_Product_Order_Service_placeOrder.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_getLastTransaction_CloudInstanceUpgrade.json",
		"SoftLayer_Virtual_Guest_Service_getPowerState.json",

		"SoftLayer_Virtual_Guest_Service_getObject.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}
