package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"

	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
)

var _ = Describe("HasVM", func() {
	var (
		vmFinder *fakevm.FakeFinder
		action   HasVM
	)

	BeforeEach(func() {
		vmFinder = &fakevm.FakeFinder{}
		action = NewHasVM(vmFinder)
	})

	Describe("Run", func() {
		It("tries to find VM with given VM CID", func() {
			_, err := action.Run(1234)
			Expect(err).ToNot(HaveOccurred())

			Expect(vmFinder.FindID).To(Equal(1234))
		})

		Context("when VM is found with given CID", func() {
			It("returns true without error", func() {
				vmFinder.FindFound = true

				found, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		Context("when VM is not found with given CID", func() {
			It("returns false without error", func() {
				found, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when VM finding fails", func() {
			It("returns error", func() {
				vmFinder.FindErr = errors.New("fake-find-err")

				found, err := action.Run(1234)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-find-err"))
				Expect(found).To(BeFalse())
			})
		})
	})
})
