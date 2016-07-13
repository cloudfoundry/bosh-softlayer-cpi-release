package vm_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"

	fakebmsclient "github.com/cloudfoundry-community/bosh-softlayer-tools/clients/fakes"
	fakevm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm/fakes"
	fakesutil "github.com/cloudfoundry/bosh-softlayer-cpi/util/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	bmsclients "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

var _ = Describe("SoftLayer_Hardware_Creator", func() {
	var (
		softLayerClient *fakeslclient.FakeSoftLayerClient
		baremetalClient *fakebmsclient.FakeBmpClient
		sshClient       *fakesutil.FakeSshClient
		vmFinder        *fakevm.FakeFinder
		agentOptions    AgentOptions
		logger          boshlog.Logger
		creator         VMCreator
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		baremetalClient = fakebmsclient.NewFakeBmpClient("fake-username", "fake-api-key", "fake-url", "fake-config-path")
		sshClient = &fakesutil.FakeSshClient{}
		agentOptions = AgentOptions{Mbus: "fake-mbus"}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		vmFinder = &fakevm.FakeFinder{}

		creator = NewBaremetalCreator(
			vmFinder,
			softLayerClient,
			baremetalClient,
			agentOptions,
			logger,
		)
		bslcommon.TIMEOUT = 2 * time.Second
		bslcommon.POLLING_INTERVAL = 1 * time.Second
	})

	Describe("#Create", func() {
		var (
			agentID    string
			stemcell   bslcstem.SoftLayerStemcell
			cloudProps VMCloudProperties
			networks   Networks
			env        Environment
		)
		t := new(bool)
		*t = true

		Context("valid arguments", func() {
			BeforeEach(func() {
				agentID = "fake-agent-id"
				stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", softLayerClient, logger)

				env = Environment{}

				vmFinder.FindVM = fakevm.NewFakeVM(1234567)
				vmFinder.FindFound = true
				vmFinder.FindErr = nil
			})

			Context("provisioning vm in baremetal server", func() {
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

					It("returns a new SoftLayerVM", func() {
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
							LocalDiskFlag:                t,
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
							Baremetal:             true,
							BaremetalStemcell:     "fake-stemcell",
							BaremetalNetbootImage: "fake-netboot-image",
						}
						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}

						baremetalClient.ProvisioningBaremetalResponse = bmsclients.CreateBaremetalsResponse{
							Status: 200,
							Data: bmsclients.TaskInfo{
								TaskId: 1234567,
							},
						}

						taskCompletedJSON := `{
 								 "status": 200,
  								 "data": {
    										"info": {
      												"status": "completed"
   											 }
  							                 }
							     }`
						taskJson := bmsclients.TaskJsonResponse{}
						err := json.Unmarshal([]byte(taskCompletedJSON), &taskJson)

						serverInfoJson := `{
 								 "status": 200,
  								 "data": {
    										"info": {
      												"id": 1234567
   											 }
  							                 }
							     }`
						serverJson := bmsclients.TaskJsonResponse{}
						err = json.Unmarshal([]byte(serverInfoJson), &serverJson)

						baremetalClient.TaskJsonResponses = []bmsclients.TaskJsonResponse{taskJson, serverJson}

						setFakeSoftlayerClientFixtures(softLayerClient)
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
							LocalDiskFlag:                t,
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
							Baremetal:             true,
							BaremetalStemcell:     "fake-stemcell",
							BaremetalNetbootImage: "fake-netboot-image",
						}

						expectedCmdResults := []string{
							"",
						}
						sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
							return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
						}

						baremetalClient.ProvisioningBaremetalResponse = bmsclients.CreateBaremetalsResponse{
							Status: 200,
							Data: bmsclients.TaskInfo{
								TaskId: 1234567,
							},
						}

						taskCompletedJSON := `{
 								 "status": 200,
  								 "data": {
    										"info": {
      												"status": "completed"
   											 }
  							                 }
							     }`
						taskJson := bmsclients.TaskJsonResponse{}
						err := json.Unmarshal([]byte(taskCompletedJSON), &taskJson)

						serverInfoJson := `{
 								 "status": 200,
  								 "data": {
    										"info": {
      												"id": 1234567
   											 }
  							                 }
							     }`
						serverJson := bmsclients.TaskJsonResponse{}
						err = json.Unmarshal([]byte(serverInfoJson), &serverJson)

						baremetalClient.TaskJsonResponses = []bmsclients.TaskJsonResponse{taskJson, serverJson}
						setFakeSoftlayerClientFixtures(softLayerClient)
						vm, err := creator.Create(agentID, stemcell, cloudProps, networks, env)
						Expect(err).ToNot(HaveOccurred())
						Expect(vm.ID()).To(Equal(1234567))
					})
				})
			})
		})

		Context("invalid arguments", func() {
			Context("missing correct VMProperties", func() {
				BeforeEach(func() {
					agentID = "fake-agent-id"
					stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", softLayerClient, logger)
					networks = Networks{}
					env = Environment{}

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

					vmFinder.FindVM = fakevm.NewFakeVM(1234567)
					vmFinder.FindFound = false
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

func setFakeSoftlayerClientFixtures(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Hardware_Service_getObject.json",
		"SoftLayer_Hardware_Service_getObject.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}
