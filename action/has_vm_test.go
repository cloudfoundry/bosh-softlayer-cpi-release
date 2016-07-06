package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"

	fakevm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm/fakes"
)

var _ = Describe("HasVM", func() {
	var (
		vmFinder *fakevm.FakeFinder
		action   HasVMAction
	)

	BeforeEach(func() {
		vmFinder = &fakevm.FakeFinder{}
		action = NewHasVM(vmFinder)
	})

	Describe("Run", func() {
		Context("when VM is found with given CID", func() {
			It("returns true without error", func() {
				vmFinder.FindFound = true
				vmFinder.FindVM = fakevm.NewFakeVM(1234)

				found, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		Context("when VM is not found with given CID", func() {
			It("returns false without error", func() {
				vmFinder.FindFound = false

				found, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when VM finding fails", func() {
			It("returns error", func() {
				vmFinder.FindFound = false
				vmFinder.FindErr = errors.New("fake-find-err")

				found, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-find-err"))
				Expect(found).To(BeFalse())
			})
		})
	})
})
