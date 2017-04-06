package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"
	fakeaction "bosh-softlayer-cpi/action/fakes"

	. "bosh-softlayer-cpi/softlayer/common"
	fakescommon "bosh-softlayer-cpi/softlayer/common/fakes"
	fakestem "bosh-softlayer-cpi/softlayer/stemcell/fakes"

	"bosh-softlayer-cpi/api"
	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

var _ = Describe("CreateVM", func() {
	var (
		fakeStemcellFinder  *fakestem.FakeStemcellFinder
		fakeStemcell        *fakestem.FakeStemcell
		fakeVmCreator       *fakescommon.FakeVMCreator
		fakeVm              *fakescommon.FakeVM
		fakeCreatorProvider *fakeaction.FakeCreatorProvider
	)

	BeforeEach(func() {
		fakeStemcellFinder = &fakestem.FakeStemcellFinder{}
		fakeVm = &fakescommon.FakeVM{}
		fakeStemcell = &fakestem.FakeStemcell{}
		fakeVmCreator = &fakescommon.FakeVMCreator{}
		fakeCreatorProvider = &fakeaction.FakeCreatorProvider{}
	})

	Describe("Run", func() {
		var (
			vmCidString   string
			stemcellCID   StemcellCID
			err           error
			networks      Networks
			diskLocality  []DiskCID
			env           Environment
			action        CreateVMAction
			fakeCloudProp VMCloudProperties
			fakeOptions   *ConcreteFactoryOptions
		)

		BeforeEach(func() {
			stemcellCID = StemcellCID(1234)
			networks = Networks{"fake-net-name": Network{IP: "fake-ip"}}
			env = Environment{"fake-env-key": "fake-env-value"}
			diskLocality = []DiskCID{1234}
		})

		JustBeforeEach(func() {
			vmCidString, err = action.Run("fake-agent-id", stemcellCID, fakeCloudProp, networks, diskLocality, env)
		})

		Context("when create vm with enabled pool succeeds", func() {
			BeforeEach(func() {
				fakeOptions = &ConcreteFactoryOptions{
					Softlayer: SoftLayerConfig{FeatureOptions: FeatureOptions{EnablePool: true}},
				}
				fakeCloudProp = VMCloudProperties{
					StartCpus:    2,
					MaxMemory:    2048,
					Datacenter:   "fake-datacenter",
					VmNamePrefix: "fake-hostname",
					BoshIp:       "10.0.0.0",
					SshKeys: []sldatatypes.SshKey{
						sldatatypes.SshKey{Id: 1234},
					},
				}
				action = NewCreateVM(fakeStemcellFinder, fakeCreatorProvider, *fakeOptions)

				fakeVm.IDReturns(1234567)
				fakeStemcellFinder.FindByIdReturns(fakeStemcell, nil)
				fakeCreatorProvider.GetReturns(fakeVmCreator)
				fakeVmCreator.CreateReturns(fakeVm, nil)

			})

			It("fetches stemcell by id", func() {
				Expect(fakeStemcellFinder.FindByIdCallCount()).To(Equal(1))
				actualId := fakeStemcellFinder.FindByIdArgsForCall(0)
				Expect(actualId).To(Equal(1234))
			})

			It("fetches creator by pool", func() {
				Expect(fakeCreatorProvider.GetCallCount()).To(Equal(1))
				actualKey := fakeCreatorProvider.GetArgsForCall(0)
				Expect(actualKey).To(Equal("pool"))
			})

			It("no error return", func() {
				Expect(fakeVmCreator.CreateCallCount()).To(Equal(1))
				actualAgentId, actualStemcell, _, _, actualEnv := fakeVmCreator.CreateArgsForCall(0)
				Expect(actualAgentId).To(Equal("fake-agent-id"))
				Expect(actualStemcell).To(Equal(fakeStemcell))
				Expect(actualEnv).To(Equal(env))
				Expect(vmCidString).To(Equal(VMCID(fakeVm.ID()).String()))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when vm name prefix is specified", func() {
			BeforeEach(func() {
				fakeOptions = &ConcreteFactoryOptions{
					Softlayer: SoftLayerConfig{FeatureOptions: FeatureOptions{EnablePool: true}},
				}
				fakeCloudProp = VMCloudProperties{
					StartCpus:    2,
					MaxMemory:    2048,
					Datacenter:   "fake-datacenter",
					VmNamePrefix: "",
					BoshIp:       "10.0.0.0",
					SshKeys: []sldatatypes.SshKey{
						sldatatypes.SshKey{Id: 1234},
					},
				}
				action = NewCreateVM(fakeStemcellFinder, fakeCreatorProvider, *fakeOptions)

				fakeVm.IDReturns(1234567)
				fakeStemcellFinder.FindByIdReturns(fakeStemcell, nil)
				fakeCreatorProvider.GetReturns(fakeVmCreator)
				fakeVmCreator.CreateReturns(fakeVm, nil)
			})
			Context("when vm name is < 64 character long", func() {
				BeforeEach(func() {
					fakeCloudProp.VmNamePrefix = "63charac_long_name_with_suffix"
				})
				It("does not modify the name length", func() {
					Expect(api.LengthOfHostName).To(Equal(63))
					Expect(fakeStemcellFinder.FindByIdCallCount()).To(Equal(1))
				})
			})
			Context("when vm name is exactly 64 character long", func() {
				BeforeEach(func() {
					fakeCloudProp.VmNamePrefix = "64charact_long_name_with_suffix"
				})
				It("adds 2 additional characters to the name", func() {
					Expect(api.LengthOfHostName).To(Equal(66))
					Expect(fakeStemcellFinder.FindByIdCallCount()).To(Equal(1))
				})
			})
			Context("when vm name > 64 characters", func() {
				BeforeEach(func() {
					fakeCloudProp.VmNamePrefix = "65characte_long_name_with_suffix"
				})
				It("does not modify the name length", func() {
					Expect(api.LengthOfHostName).To(Equal(65))
					Expect(fakeStemcellFinder.FindByIdCallCount()).To(Equal(1))
				})
			})
		})

		Context("when stemcell not found", func() {
			BeforeEach(func() {
				fakeOptions = &ConcreteFactoryOptions{
					Softlayer: SoftLayerConfig{FeatureOptions: FeatureOptions{EnablePool: true}},
				}
				action = NewCreateVM(fakeStemcellFinder, fakeCreatorProvider, *fakeOptions)

				fakeStemcellFinder.FindByIdReturns(nil, errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err.Error()).To(ContainSubstring("kaboom"))
				Expect(vmCidString).To(Equal("0"))
			})
		})

		Context("when create vm with enabled pool error out", func() {
			BeforeEach(func() {
				fakeOptions = &ConcreteFactoryOptions{
					Softlayer: SoftLayerConfig{FeatureOptions: FeatureOptions{EnablePool: true}},
				}
				action = NewCreateVM(fakeStemcellFinder, fakeCreatorProvider, *fakeOptions)

				fakeCreatorProvider.GetReturns(fakeVmCreator)
				fakeStemcellFinder.FindByIdReturns(fakeStemcell, nil)
				fakeVmCreator.CreateReturns(nil, errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err.Error()).To(ContainSubstring("kaboom"))
				Expect(vmCidString).To(Equal("0"))
			})
		})

		Context("when create vm without pool", func() {
			BeforeEach(func() {
				fakeOptions = &ConcreteFactoryOptions{
					Softlayer: SoftLayerConfig{FeatureOptions: FeatureOptions{EnablePool: false}},
				}
				fakeCloudProp = VMCloudProperties{
					StartCpus:    2,
					MaxMemory:    2048,
					Datacenter:   "fake-datacenter",
					VmNamePrefix: "fake-hostname",
					BoshIp:       "10.0.0.0",
					SshKeys: []sldatatypes.SshKey{
						{Id: 1234},
					},
				}
				action = NewCreateVM(fakeStemcellFinder, fakeCreatorProvider, *fakeOptions)

				fakeVm.IDReturns(1234567)
				fakeStemcellFinder.FindByIdReturns(fakeStemcell, nil)
				fakeCreatorProvider.GetReturns(fakeVmCreator)
				fakeVmCreator.CreateReturns(fakeVm, nil)
			})

			It("fetches creator by virtualguest", func() {
				Expect(fakeCreatorProvider.GetCallCount()).To(Equal(1))
				actualKey := fakeCreatorProvider.GetArgsForCall(0)
				Expect(actualKey).To(Equal("virtualguest"))
			})
		})

		Context("when create baremetal", func() {
			BeforeEach(func() {
				fakeOptions = &ConcreteFactoryOptions{
					Softlayer: SoftLayerConfig{FeatureOptions: FeatureOptions{EnablePool: false}},
				}
				fakeCloudProp = VMCloudProperties{
					StartCpus:    2,
					MaxMemory:    2048,
					Datacenter:   "fake-datacenter",
					VmNamePrefix: "fake-hostname",
					Baremetal:    true,
					BoshIp:       "10.0.0.0",
					SshKeys: []sldatatypes.SshKey{
						{Id: 1234},
					},
				}
				action = NewCreateVM(fakeStemcellFinder, fakeCreatorProvider, *fakeOptions)

				fakeVm.IDReturns(1234567)
				fakeStemcellFinder.FindByIdReturns(fakeStemcell, nil)
				fakeCreatorProvider.GetReturns(fakeVmCreator)
				fakeVmCreator.CreateReturns(fakeVm, nil)
			})

			It("fetches creator by baremetal", func() {
				Expect(fakeCreatorProvider.GetCallCount()).To(Equal(1))
				actualKey := fakeCreatorProvider.GetArgsForCall(0)
				Expect(actualKey).To(Equal("baremetal"))
			})
		})
	})
})
