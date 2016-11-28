package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"

	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"
)

var _ = Describe("RebootVM", func() {
	var (
		fakeVmFinder *fakescommon.FakeVMFinder
		fakeVm       *fakescommon.FakeVM

		action RebootVMAction
	)

	BeforeEach(func() {
		fakeVmFinder = &fakescommon.FakeVMFinder{}
		fakeVm = &fakescommon.FakeVM{}
		action = NewRebootVM(fakeVmFinder)
	})

	Describe("Run", func() {
		var (
			vmCid VMCID
			err   error
		)

		BeforeEach(func() {
			vmCid = VMCID(123456)
		})

		JustBeforeEach(func() {
			_, err = action.Run(vmCid)
		})

		Context("when reboot vm succeeds", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(fakeVm, true, nil)
			})

			It("fetches vm by cid", func() {
				Expect(fakeVmFinder.FindCallCount()).To(Equal(1))
				actualCid := fakeVmFinder.FindArgsForCall(0)
				Expect(actualCid).To(Equal(123456))
			})

			It("no error return", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when find vm error out", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(nil, false, errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})

		Context("when find vm return false", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(nil, false, nil)
			})

			It("no error return", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when reboot vm error out", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeVm.RebootReturns(errors.New("kaboom"))
			})
			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})
	})
})
