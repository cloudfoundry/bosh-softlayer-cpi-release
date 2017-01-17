package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"

	"fmt"
	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"
	fakedisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk/fakes"
)

var _ = Describe("AttachDisk", func() {
	var (
		fakeVmFinder   *fakescommon.FakeVMFinder
		fakeVm         *fakescommon.FakeVM
		fakeDiskFinder *fakedisk.FakeDiskFinder
		fakeDisk       *fakedisk.FakeDisk
		action         AttachDiskAction
	)

	BeforeEach(func() {
		fakeVmFinder = &fakescommon.FakeVMFinder{}
		fakeVm = &fakescommon.FakeVM{}
		fakeDiskFinder = &fakedisk.FakeDiskFinder{}
		fakeDisk = &fakedisk.FakeDisk{}
		action = NewAttachDisk(fakeVmFinder, fakeDiskFinder)
	})

	Describe("Run", func() {
		var (
			vmCid   VMCID
			diskCID DiskCID

			err error
		)

		BeforeEach(func() {
			vmCid = VMCID(123456)
			diskCID = DiskCID(123456)
		})

		JustBeforeEach(func() {
			_, err = action.Run(vmCid, diskCID)
		})

		Context("when attach disk succeeds", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeDiskFinder.FindReturns(fakeDisk, true, nil)

				fakeVm.AttachDiskReturns(nil)
			})

			It("fetches vm by cid", func() {
				Expect(fakeVmFinder.FindCallCount()).To(Equal(1))
				actualCid := fakeVmFinder.FindArgsForCall(0)
				Expect(actualCid).To(Equal(123456))
			})

			It("fetches disk by cid", func() {
				Expect(fakeDiskFinder.FindCallCount()).To(Equal(1))
				actualCid := fakeDiskFinder.FindArgsForCall(0)
				Expect(actualCid).To(Equal(123456))
			})

			It("no error return", func() {
				Expect(fakeVm.AttachDiskCallCount()).To(Equal(1))
				actualDisk := fakeVm.AttachDiskArgsForCall(0)
				Expect(actualDisk).To(Equal(fakeDisk))
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

			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Sprintf("Expected to find VM '%s'", vmCid)))
			})
		})

		Context("when find disk error out", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeDiskFinder.FindReturns(nil, false, errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})

		Context("when find disk return false", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeDiskFinder.FindReturns(nil, false, nil)
			})

			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Expected to find disk '%s'", diskCID)))
			})
		})

		Context("when attach disk error out", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeDiskFinder.FindReturns(fakeDisk, true, nil)

				fakeVm.AttachDiskReturns(errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})
	})
})
