package action_test

import (
	. "github.com/maximilien/bosh-softlayer-cpi/action"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
)

var _ = Describe("SetVMMetadata", func() {
	var (
		vmFinder *fakevm.FakeFinder
		action   SetVMMetadata
	)

	BeforeEach(func() {
		vmFinder = &fakevm.FakeFinder{}
		action = NewSetVMMetadata(vmFinder)
	})

	Describe("Run", func() {
		It("does not do anything and return no error", func() {
			_, err := action.Run(1234, bslcvm.VMMetadata{})
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
