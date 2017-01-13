package pool_test

import (
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"

	fakespool "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/client/vm/fakes"

	"github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/client/vm"
	"github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/models"

	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	fakesutil "github.com/cloudfoundry/bosh-softlayer-cpi/util/fakes"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

var _ = Describe("SoftlayerPoolCreator", func() {
	var (
		softLayerClient *fakeslclient.FakeSoftLayerClient
		fakePoolClient  *fakespool.FakeSoftLayerPoolClient
		sshClient       *fakesutil.FakeSshClient
		fakeVmFinder    *fakescommon.FakeVMFinder
		agentOptions    AgentOptions
		logger          boshlog.Logger
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		fakePoolClient = &fakespool.FakeSoftLayerPoolClient{}
		sshClient = &fakesutil.FakeSshClient{}
		agentOptions = AgentOptions{Mbus: "fake-mbus"}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fakeVmFinder = &fakescommon.FakeVMFinder{}
	})

	Describe("create from pool", func() {
		var (
			err        error
			agentID    string
			stemcell   bslcstem.SoftLayerStemcell
			cloudProps VMCloudProperties
			networks   Networks
			env        Environment

			creator        VMCreator
			featureOptions *FeatureOptions

			fakeVm         *fakescommon.FakeVM
			poolVmResponse *models.VMResponse
			poolVm         *models.VM

			expectedCmdResults []string

			actualVm VM
		)

		BeforeEach(func() {
			fakeVm = &fakescommon.FakeVM{}

			poolVm = &models.VM{
				Cid:         int32(1234567),
				CPU:         int32(4),
				MemoryMb:    int32(2048),
				PrivateVlan: int32(524956),
				PublicVlan:  int32(524956),
			}

			poolVmResponse = &models.VMResponse{
				VM: poolVm,
			}

			agentID = "fake-agent-id"
			stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", softLayerClient, logger)
			env = Environment{}
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

			featureOptions = &FeatureOptions{
				EnablePool: true,
			}
			creator = NewSoftLayerPoolCreator(fakeVmFinder, fakePoolClient, softLayerClient, agentOptions, logger, *featureOptions)

			expectedCmdResults = []string{
				"",
			}

			sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
				return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
			}
		})

		JustBeforeEach(func() {
			actualVm, err = creator.Create(agentID, stemcell, cloudProps, networks, env)
		})

		Context("when free vm in pool, create vm succeeds", func() {
			BeforeEach(func() {
				setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize_OS_Reload(softLayerClient)

				fakeVm.IDReturns(1234567)
				fakePoolClient.OrderVMByFilterReturns(vm.NewOrderVMByFilterOK().WithPayload(poolVmResponse), nil)
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeVm.ConfigureNetworks2Returns(nil)
				fakeVm.UpdateAgentEnvReturns(nil)

				//fakePoolClient.UpdateVMWithStateReturns(vm.NewUpdateVMWithStateOK().WithPayload("updated successfully"), nil)
				fakePoolClient.UpdateVMReturns(vm.NewUpdateVMOK(), nil)
			})

			It("order vm by filter", func() {
				Expect(fakePoolClient.OrderVMByFilterCallCount()).To(Equal(1))
				orderVMByFilterParams := fakePoolClient.OrderVMByFilterArgsForCall(0)
				Expect(orderVMByFilterParams.Body.CPU).To(Equal(int32(4)))
				Expect(orderVMByFilterParams.Body.PublicVlan).To(Equal(int32(524956)))
			})
			It("update vm state to using", func() {
				Expect(fakePoolClient.UpdateVMCallCount()).To(Equal(1))
				updateVMParams := fakePoolClient.UpdateVMArgsForCall(0)
				Expect(updateVMParams.Body.State).To(Equal(models.StateUsing))
			})

			It("no error return", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(actualVm.ID()).To(Equal(1234567))
			})
		})

		Context("when no free vm in pool, create succeeds", func() {
			BeforeEach(func() {
				setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize(softLayerClient)

				fakeVm.IDReturns(1234567)
				fakePoolClient.OrderVMByFilterReturns(nil, vm.NewOrderVMByFilterNotFound())
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeVm.ConfigureNetworks2Returns(nil)
				fakeVm.UpdateAgentEnvReturns(nil)

				fakePoolClient.AddVMReturns(vm.NewAddVMOK().WithPayload("added successfully"), nil)
			})

			It("order vm by filter", func() {
				Expect(fakePoolClient.OrderVMByFilterCallCount()).To(Equal(1))
				orderVMByFilterParams := fakePoolClient.OrderVMByFilterArgsForCall(0)
				Expect(orderVMByFilterParams.Body.CPU).To(Equal(int32(4)))
				Expect(orderVMByFilterParams.Body.PublicVlan).To(Equal(int32(524956)))
			})
			It("add vm to pool after creating in softlayer", func() {
				Expect(fakePoolClient.AddVMCallCount()).To(Equal(1))
				addVMParams := fakePoolClient.AddVMArgsForCall(0)
				Expect(addVMParams.Body.Cid).To(Equal(int32(1234567)))
			})

			It("no error return", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(actualVm.ID()).To(Equal(1234567))
			})
		})

		Context("when order vm by filter from pool error out", func() {
			BeforeEach(func() {
				fakePoolClient.OrderVMByFilterReturns(nil, vm.NewOrderVMByFilterDefault(500))
			})

			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Ordering vm from pool"))
			})
		})

		Context("when add vm to pool error out", func() {
			BeforeEach(func() {
				setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize(softLayerClient)

				fakeVm.IDReturns(1234567)
				fakePoolClient.OrderVMByFilterReturns(nil, vm.NewOrderVMByFilterNotFound())
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeVm.ConfigureNetworks2Returns(nil)
				fakeVm.UpdateAgentEnvReturns(nil)

				fakePoolClient.AddVMReturns(nil, vm.NewAddVMDefault(500))
			})

			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Adding vm into pool"))
			})
		})

		Context("when update vm to using error out", func() {
			BeforeEach(func() {
				setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize_OS_Reload(softLayerClient)

				fakeVm.IDReturns(1234567)
				fakePoolClient.OrderVMByFilterReturns(vm.NewOrderVMByFilterOK().WithPayload(poolVmResponse), nil)
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeVm.ConfigureNetworks2Returns(nil)
				fakeVm.UpdateAgentEnvReturns(nil)

				//fakePoolClient.UpdateVMWithStateReturns(nil, vm.NewUpdateVMDefault(500))
				fakePoolClient.UpdateVMReturns(nil, vm.NewUpdateVMDefault(500))
			})

			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Updating the hostname of vm"))
			})
		})

		Context("when doing os_reload directly without operation of pool succeeds", func() {
			BeforeEach(func() {
				networks = map[string]Network{
					"fake-network0": Network{
						Type:    "dynamic",
						IP:      "10.0.0.1",
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

				setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize_OS_Reload_2(softLayerClient)

				fakeVm.IDReturns(1234567)
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeVm.ConfigureNetworks2Returns(nil)
				fakeVm.UpdateAgentEnvReturns(nil)
			})

			It("no error return", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

func setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize_OS_Reload(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
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

func setFakeSoftlayerClientCreateObjectTestFixturesWithEphemeralDiskSize_OS_Reload_2(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
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
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}

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
