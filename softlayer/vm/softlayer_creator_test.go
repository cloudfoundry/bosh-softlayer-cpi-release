package vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	testhelpers "github.com/maximilien/bosh-softlayer-cpi/test_helpers"

	fakestem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell/fakes"
	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	bslcommon "github.com/maximilien/bosh-softlayer-cpi/softlayer/common"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

var _ = Describe("SoftLayerCreator", func() {
	var (
		softLayerClient        *fakeslclient.FakeSoftLayerClient
		agentEnvServiceFactory *fakevm.FakeAgentEnvServiceFactory
		agentOptions           AgentOptions
		logger                 boshlog.Logger
		creator                SoftLayerCreator
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")

		agentEnvServiceFactory = &fakevm.FakeAgentEnvServiceFactory{}
		agentOptions = AgentOptions{Mbus: "fake-mbus"}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		creator = NewSoftLayerCreator(
			softLayerClient,
			agentEnvServiceFactory,
			agentOptions,
			logger,
		)
		bslcommon.TIMEOUT = 2 * time.Second
		bslcommon.POLLING_INTERVAL = 1 * time.Second
		bslcommon.PAUSE_TIME = 1 * time.Second
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
					UserData: []sldatatypes.UserData{{Value: "fake-userdata"}},
				}
				networks = Networks{}
				env = Environment{}

				setFakeSoftLayerClientCreateObjectTestFixtures(softLayerClient)
			})

			It("returns a new SoftLayerVM with correct virtual guest ID and SoftLayerClient", func() {
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

					setFakeSoftLayerClientCreateObjectTestFixtures(softLayerClient)
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

				It("fails when Ephemeral DiskSize is negative", func() {
					cloudProps = VMCloudProperties{
						StartCpus:         4,
						MaxMemory:         1024,
						Domain:            "fake-domain",
						Datacenter:        sldatatypes.Datacenter{Name: "fake-datacenter"},
						RootDiskSize:      100,
						EphemeralDiskSize: -100,
					}

					_, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})

func setFakeSoftLayerClientCreateObjectTestFixtures(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_createObject.json",
		"SoftLayer_Virtual_Guest_Service_getPowerState.json",

		"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",

		"SoftLayer_Virtual_Guest_Service_setMetadata.json",
		"SoftLayer_Virtual_Guest_Service_configureMetadataDisk.json",

		"SoftLayer_Virtual_Guest_Service_getPowerState.json",

		"SoftLayer_Virtual_Guest_Service_getPowerState.json",
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",
		"SoftLayer_Virtual_Guest_Service_getUpgradeItemPrices.json",
		"SoftLayer_Product_Order_Service_placeOrder.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}
