package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
)

var _ = Describe("SetVMMetadata", func() {
	var (
		vmFinder *fakevm.FakeFinder
		action   SetVMMetadata
		metadata bslcvm.VMMetadata
	)

	BeforeEach(func() {
		vmFinder = &fakevm.FakeFinder{}
		action = NewSetVMMetadata(vmFinder)
		metadata = bslcvm.VMMetadata{}
	})

	Describe("Run", func() {
		It("tries to find vm with given vm cid", func() {
			_, err := action.Run(1234, metadata)
			Expect(err).ToNot(HaveOccurred())

			Expect(vmFinder.FindID).To(Equal(1234))
		})

		Context("when vm is found with given vm cid", func() {
			var (
				vm *fakevm.FakeVM
			)

			BeforeEach(func() {
				vm = fakevm.NewFakeVM(1234)
				vmFinder.FindVM = vm
				vmFinder.FindFound = true
			})

			It("sets vm metadata", func() {
				_, err := action.Run(1234, metadata)
				Expect(err).ToNot(HaveOccurred())

				Expect(vm.SetMetadataCalled).To(BeTrue())
				Expect(vm.VMMetadata).To(Equal(metadata))
			})

			It("returns error if setting metadata fails", func() {
				vm.SetMetadataErr = errors.New("fake-set-vm-metadata-err")

				_, err := action.Run(1234, metadata)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-set-vm-metadata-err"))
			})
		})

		Context("when vm is not found with given cid", func() {
			It("does vmFinder return error", func() {
				vmFinder.FindFound = false

				_, err := action.Run(1234, metadata)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when vm finding fails", func() {
			It("does not return error", func() {
				vmFinder.FindErr = errors.New("fake-find-err")

				_, err := action.Run(1234, metadata)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-find-err"))
			})
		})
	})
})
