package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"

	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
)

var _ = Describe("RebootVM", func() {
	var (
		vmFinder *fakevm.FakeFinder
		action   RebootVM
	)

	BeforeEach(func() {
		vmFinder = &fakevm.FakeFinder{}
		action = NewRebootVM(vmFinder)
	})

	Describe("Run", func() {
		It("tries to find vm with given vm cid", func() {
			_, err := action.Run(1234)
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

			It("reboots vm", func() {
				_, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())

				Expect(vm.RebootCalled).To(BeTrue())
			})

			It("returns error if rebooting vm fails", func() {
				vm.RebootErr = errors.New("fake-reboot-err")

				_, err := action.Run(1234)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-reboot-err"))
			})
		})

		Context("when vm is not found with given cid", func() {
			It("does vmFinder return error", func() {
				vmFinder.FindFound = false

				_, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when vm finding fails", func() {
			It("does not return error", func() {
				vmFinder.FindErr = errors.New("fake-find-err")

				_, err := action.Run(1234)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-find-err"))
			})
		})
	})
})
