package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"

	fakestem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell/fakes"
	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

var _ = Describe("CreateVM", func() {
	var (
		stemcellFinder *fakestem.FakeFinder
		vmCreator      *fakevm.FakeCreator
		action         CreateVM
	)

	BeforeEach(func() {
		stemcellFinder = &fakestem.FakeFinder{}
		vmCreator = &fakevm.FakeCreator{}
		action = NewCreateVM(stemcellFinder, vmCreator)
	})

	Describe("Run", func() {
		var (
			stemcellCID               StemcellCID
			vmCloudProp, vmCloudProp2 bslcvm.VMCloudProperties
			networks                  Networks
			diskLocality              []DiskCID
			env                       Environment
		)

		BeforeEach(func() {
			stemcellCID = StemcellCID(1234)
			vmCloudProp = bslcvm.VMCloudProperties{
				StartCpus:  2,
				MaxMemory:  2048,
				Datacenter: sldatatypes.Datacenter{Name: "fake-datacenter"},
				SshKeys: []sldatatypes.SshKey{
					sldatatypes.SshKey{Id: 1234},
				},
			}
			networks = Networks{"fake-net-name": Network{IP: "fake-ip"}}
			diskLocality = []DiskCID{1234}
			env = Environment{"fake-env-key": "fake-env-value"}
		})

		It("tries to find stemcell with given stemcell cid", func() {
			stemcellFinder.FindFound = true
			vmCreator.CreateVM = fakevm.NewFakeVM(1234)

			_, err := action.Run("fake-agent-id", stemcellCID, vmCloudProp, networks, diskLocality, env)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when stemcell is found with given stemcell cid", func() {
			var (
				stemcell *fakestem.FakeStemcell
			)

			BeforeEach(func() {
				stemcell = fakestem.NewFakeStemcell(1234, "fake-stemcell-id", fakestem.FakeStemcellKind)
				stemcellFinder.FindStemcell = stemcell
				stemcellFinder.FindFound = true
			})

			It("returns id for created VM", func() {
				vmCreator.CreateVM = fakevm.NewFakeVM(1234)

				id, err := action.Run("fake-agent-id", stemcellCID, vmCloudProp, networks, diskLocality, env)
				Expect(err).ToNot(HaveOccurred())
				Expect(id).To(Equal(VMCID(1234).String()))
			})

			It("creates VM with requested agent ID, stemcell, cloud properties, and networks", func() {
				vmCreator.CreateVM = fakevm.NewFakeVM(1234)

				_, err := action.Run("fake-agent-id", stemcellCID, vmCloudProp, networks, diskLocality, env)
				Expect(err).ToNot(HaveOccurred())

				Expect(vmCreator.CreateAgentID).To(Equal("fake-agent-id"))
				Expect(vmCreator.CreateStemcell).To(Equal(stemcell))
				Expect(vmCreator.CreateNetworks).To(Equal(networks.AsVMNetworks()))
				Expect(vmCreator.CreateEnvironment).To(Equal(
					bslcvm.Environment{"fake-env-key": "fake-env-value"},
				))
			})

			It("creates VM with requested agent ID, stemcell, cloud properties (without startCPU, Memory, NetworkSpeed), and networks", func() {
				vmCloudProp2 = bslcvm.VMCloudProperties{
					Datacenter: sldatatypes.Datacenter{Name: "fake-datacenter"},
					SshKeys: []sldatatypes.SshKey{
						sldatatypes.SshKey{Id: 1234},
					},
				}
				vmCreator.CreateVM = fakevm.NewFakeVM(1234)

				_, err := action.Run("fake-agent-id", stemcellCID, vmCloudProp2, networks, diskLocality, env)
				Expect(err).ToNot(HaveOccurred())

				Expect(vmCreator.CreateAgentID).To(Equal("fake-agent-id"))
				Expect(vmCreator.CreateStemcell).To(Equal(stemcell))
				Expect(vmCreator.CreateNetworks).To(Equal(networks.AsVMNetworks()))
				Expect(vmCreator.CreateEnvironment).To(Equal(
					bslcvm.Environment{"fake-env-key": "fake-env-value"},
				))
			})

			It("returns error if creating VM fails", func() {
				vmCreator.CreateErr = errors.New("fake-create-err")

				id, err := action.Run("fake-agent-id", stemcellCID, vmCloudProp, networks, diskLocality, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-err"))
				Expect(id).To(Equal(VMCID(0).String()))
			})
		})

		Context("when stemcell is not found with given cid", func() {
			It("returns error because VM cannot be created without a stemcell", func() {
				stemcellFinder.FindFound = false

				id, err := action.Run("fake-agent-id", stemcellCID, vmCloudProp, networks, diskLocality, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Expected to find stemcell"))
				Expect(id).To(Equal(VMCID(0).String()))
			})
		})

		Context("when stemcell finding fails", func() {
			It("returns error because VM cannot be created without a stemcell", func() {
				stemcellFinder.FindErr = errors.New("fake-find-err")

				id, err := action.Run("fake-agent-id", stemcellCID, vmCloudProp, networks, diskLocality, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-find-err"))
				Expect(id).To(Equal(VMCID(0).String()))
			})
		})
	})
})
