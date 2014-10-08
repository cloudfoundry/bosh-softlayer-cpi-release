package vm_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	common "github.com/maximilien/bosh-softlayer-cpi/common"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

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
		workingDir, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())

		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		softLayerClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures(filepath.Join(workingDir, "..", ".."), "softlayer", "SoftLayer_Virtual_Guest_Service_createObject.json")
		Expect(err).ToNot(HaveOccurred())

		agentEnvServiceFactory = &fakevm.FakeAgentEnvServiceFactory{}
		agentOptions = AgentOptions{Mbus: "fake-mbus"}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		creator = NewSoftLayerCreator(
			softLayerClient,
			agentEnvServiceFactory,
			agentOptions,
			logger,
		)
	})

	Describe("Create", func() {
		var (
			agentID    string
			stemcell   bslcstem.FSStemcell
			cloudProps VMCloudProperties
			networks   Networks
			env        Environment
		)

		Context("valid arguments", func() {
			BeforeEach(func() {
				agentID = "fake-agent-id"
				stemcell = bslcstem.NewFSStemcell("fake-stemcell-id", logger)
				cloudProps = VMCloudProperties{
					StartCpus:  4,
					MaxMemory:  2048,
					Datacenter: sldatatypes.Datacenter{Name: "fake-datacenter"},
				}
				networks = Networks{}
				env = Environment{}
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
					stemcell = bslcstem.NewFSStemcell("fake-stemcell-id", logger)
					networks = Networks{}
					env = Environment{}
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

				It("fails when VMProperties is missing MaxMemory", func() {
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
