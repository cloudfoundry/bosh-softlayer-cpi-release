package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"

	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"
)

var _ = Describe("HasVM", func() {
	var (
		fakeVmFinder *fakescommon.FakeVMFinder
		action       HasVMAction
	)

	BeforeEach(func() {
		fakeVmFinder = &fakescommon.FakeVMFinder{}
		action = NewHasVM(fakeVmFinder)
	})

	Describe("Run", func() {
		var (
			vmCid VMCID
			found bool
			err   error
		)
		BeforeEach(func() {
			vmCid = VMCID(123456)
		})

		JustBeforeEach(func() {
			found, err = action.Run(vmCid)
		})
		Context("when has vm succeeds", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(nil, true, nil)
			})
			It("no error return", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		Context("when has vm fails", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(nil, false, errors.New("kaboom"))
			})
			It("no error return", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
